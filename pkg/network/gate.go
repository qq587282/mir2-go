package network

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/mir2go/mir2/pkg/protocol"
	"go.uber.org/zap"
)

type GateSession struct {
	Conn         net.Conn
	Addr         string
	SessionID    int32
	Buffer       []byte
	ReceiveTick  time.Time
	SendTick     time.Time
	Alive        bool
	UserData     interface{}
	SendChan     chan []byte
	ReceiveMutex sync.Mutex
	SendMutex   sync.Mutex
}

type GateServer struct {
	Addr        string
	Listener    net.Listener
	SessionList map[int32]*GateSession
	SessionMutex sync.RWMutex
	NextSessionID int32
	MaxSessions int
	Logger      *zap.Logger
	OnConnect   func(*GateSession)
	OnDisconnect func(*GateSession)
	OnMessage   func(*GateSession, []byte)
	Running     bool
}

func NewGateServer(addr string, logger *zap.Logger) *GateServer {
	return &GateServer{
		Addr:        addr,
		SessionList: make(map[int32]*GateSession),
		NextSessionID: 1,
		MaxSessions: 10000,
		Logger:      logger,
	}
}

func (g *GateServer) Start() error {
	var err error
	g.Listener, err = net.Listen("tcp", g.Addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	g.Running = true
	g.Logger.Info("Gate server started", zap.String("addr", g.Addr))
	
	go g.acceptLoop()
	return nil
}

func (g *GateServer) Stop() {
	g.Running = false
	if g.Listener != nil {
		g.Listener.Close()
	}
	
	g.SessionMutex.Lock()
	defer g.SessionMutex.Unlock()
	for _, sess := range g.SessionList {
		if sess.Conn != nil {
			sess.Conn.Close()
		}
	}
	g.SessionList = make(map[int32]*GateSession)
	g.Logger.Info("Gate server stopped")
}

func (g *GateServer) acceptLoop() {
	for g.Running {
		conn, err := g.Listener.Accept()
		if err != nil {
			if g.Running {
				g.Logger.Error("accept error", zap.Error(err))
			}
			continue
		}
		
		g.SessionMutex.Lock()
		if len(g.SessionList) >= g.MaxSessions {
			g.SessionMutex.Unlock()
			conn.Close()
			continue
		}
		
		sess := &GateSession{
			Conn:        conn,
			Addr:        conn.RemoteAddr().String(),
			SessionID:   g.NextSessionID,
			Buffer:      make([]byte, 0, 16384),
			ReceiveTick: time.Now(),
			SendTick:    time.Now(),
			Alive:       true,
			SendChan:    make(chan []byte, 100),
		}
		g.NextSessionID++
		g.SessionList[sess.SessionID] = sess
		g.SessionMutex.Unlock()
		
		g.Logger.Info("new connection", zap.String("addr", sess.Addr), zap.Int32("session", sess.SessionID))
		
		if g.OnConnect != nil {
			g.OnConnect(sess)
		}
		
		go g.sessionLoop(sess)
	}
}

func (g *GateServer) sessionLoop(sess *GateSession) {
	defer func() {
		if r := recover(); r != nil {
			g.Logger.Error("session panic", zap.Any("error", r), zap.Int32("session", sess.SessionID))
		}
		
		g.SessionMutex.Lock()
		delete(g.SessionList, sess.SessionID)
		g.SessionMutex.Unlock()
		
		if sess.Conn != nil {
			sess.Conn.Close()
		}
		
		if g.OnDisconnect != nil {
			g.OnDisconnect(sess)
		}
		g.Logger.Info("session closed", zap.String("addr", sess.Addr), zap.Int32("session", sess.SessionID))
	}()
	
	conn := sess.Conn
	conn.SetReadDeadline(time.Now().Add(time.Minute * 30))
	conn.SetWriteDeadline(time.Now().Add(time.Second * 30))
	
	go sess.sendLoop()
	
	buf := make([]byte, 8192)
	for sess.Alive {
		n, err := conn.Read(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			if err != io.EOF {
				g.Logger.Error("read error", zap.Error(err))
			}
			break
		}
		
		sess.ReceiveMutex.Lock()
		sess.Buffer = append(sess.Buffer, buf[:n]...)
		sess.ReceiveTick = time.Now()
		sess.ReceiveMutex.Unlock()
		
		g.processBuffer(sess)
	}
}

func (g *GateServer) processBuffer(sess *GateSession) {
	sess.ReceiveMutex.Lock()
	defer sess.ReceiveMutex.Unlock()
	
	buf := sess.Buffer
	for len(buf) >= protocol.DEFBLOCKSIZE {
		header := binary.LittleEndian.Uint32(buf[0:4])
		if header != protocol.RUNGATECODE {
			if len(buf) > 1 {
				buf = buf[1:]
				sess.Buffer = buf
			}
			continue
		}
		
		if len(buf) < 4 {
			break
		}
		
		length := binary.LittleEndian.Uint16(buf[2:4])
		totalLen := 4 + int(length)
		
		if len(buf) < totalLen {
			break
		}
		
		packet := make([]byte, length)
		copy(packet, buf[4:totalLen])
		
		if g.OnMessage != nil {
			g.OnMessage(sess, packet)
		}
		
		buf = buf[totalLen:]
	}
	
	sess.Buffer = buf
}

func (s *GateSession) sendLoop() {
	defer s.Conn.Close()
	
	for data := range s.SendChan {
		s.SendMutex.Lock()
		_, err := s.Conn.Write(data)
		s.SendMutex.Unlock()
		
		if err != nil {
			s.Alive = false
			return
		}
		s.SendTick = time.Now()
	}
}

func (s *GateSession) Send(packet []byte) {
	if !s.Alive {
		return
	}
	
	select {
	case s.SendChan <- packet:
	default:
	}
}

func (s *GateSession) Close() {
	s.Alive = false
	close(s.SendChan)
	if s.Conn != nil {
		s.Conn.Close()
	}
}

func (g *GateServer) Broadcast(packet []byte) {
	g.SessionMutex.RLock()
	defer g.SessionMutex.RUnlock()
	
	for _, sess := range g.SessionList {
		if sess.Alive {
			sess.Send(packet)
		}
	}
}

func (g *GateServer) GetSessionCount() int {
	g.SessionMutex.RLock()
	defer g.SessionMutex.RUnlock()
	return len(g.SessionList)
}

func EncodePacket(data []byte) []byte {
	length := uint16(len(data))
	packet := make([]byte, 4+length)
	binary.LittleEndian.PutUint32(packet[0:4], protocol.RUNGATECODE)
	binary.LittleEndian.PutUint16(packet[2:4], length)
	copy(packet[4:], data)
	return packet
}

func DecodeMessages(buffer []byte) [][]byte {
	var messages [][]byte
	pos := 0
	
	for pos+14 <= len(buffer) {
		msgLen := binary.LittleEndian.Uint16(buffer[pos+12 : pos+14])
		totalLen := 14 + int(msgLen)
		
		if pos+totalLen > len(buffer) {
			break
		}
		
		msg := make([]byte, msgLen)
		copy(msg, buffer[pos+14:pos+totalLen])
		messages = append(messages, msg)
		pos += totalLen
	}
	
	return messages
}
