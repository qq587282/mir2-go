package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/mir2go/mir2/pkg/config"
	"github.com/mir2go/mir2/pkg/network"
	"github.com/mir2go/mir2/pkg/protocol"
)

var (
	logger     *zap.Logger
	server     *network.GateServer
	configFile string
	m2Addr     string
)

func init() {
	flag.StringVar(&configFile, "config", "config.yaml", "config file path")
}

func main() {
	flag.Parse()

	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.OutputPaths = []string{"stdout", "rungate.log"}
	zapConfig.ErrorOutputPaths = []string{"stderr", "rungate.log"}
	zapLogger, _ := zapConfig.Build()
	logger = zapLogger
	defer zapLogger.Sync()

	fmt.Println("!!! Starting RunGate !!!")
	logger.Info("Starting RunGate...")

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		logger.Warn("Failed to load config, using default", zap.Error(err))
		cfg = config.GetDefaultConfig()
	}

	m2Addr = fmt.Sprintf("127.0.0.1:%d", cfg.M2Server.Port)

	addr := fmt.Sprintf("%s:%d", cfg.RunGate.IP, cfg.RunGate.Port)
	server = network.NewGateServer(addr, logger)
	server.MaxSessions = cfg.RunGate.MaxConn
	server.OnConnect = onConnect
	server.OnDisconnect = onDisconnect
	server.OnMessage = onRunMessage

	if err := server.Start(); err != nil {
		logger.Error("Failed to start RunGate", zap.Error(err))
		os.Exit(1)
	}

	fmt.Println("!!! RunGate started on", addr, "!!!")
	logger.Info("RunGate started",
		zap.String("addr", addr),
		zap.Int("maxconn", cfg.RunGate.MaxConn),
		zap.String("m2server", m2Addr),
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down RunGate...")
	server.Stop()
}

type ClientSession struct {
	m2Conn net.Conn
}

var (
	clientSessions = make(map[int32]*ClientSession)
	sessionsMutex  sync.RWMutex
)

func onConnect(sess *network.GateSession) {
	fmt.Printf("!!! Player connected: %s session=%d !!!\n", sess.Addr, sess.SessionID)
	logger.Info("Player connected to RunGate",
		zap.String("addr", sess.Addr),
		zap.Int32("session", sess.SessionID),
	)

	cs := &ClientSession{}
	sessionsMutex.Lock()
	clientSessions[sess.SessionID] = cs
	sessionsMutex.Unlock()

	fmt.Printf("!!! Launching goroutine to connect to M2Server for session %d !!!\n", sess.SessionID)
	go connectM2ForClient(sess, cs)

	fmt.Printf("!!! Waiting for M2Server connection for session %d !!!\n", sess.SessionID)
}

func connectM2ForClient(sess *network.GateSession, cs *ClientSession) {
	for {
		fmt.Printf("!!! Connecting session %d to M2Server at %s !!!\n", sess.SessionID, m2Addr)
		conn, err := net.DialTimeout("tcp", m2Addr, 10 * time.Second)
		if err != nil {
			fmt.Printf("!!! M2Server connection FAILED for session %d: %v, will retry in 3s !!!\n", sess.SessionID, err)
			time.Sleep(3 * time.Second)
			continue
		}

		sessionsMutex.Lock()
		cs.m2Conn = conn
		sessionsMutex.Unlock()
		
		fmt.Printf("!!! Session %d connected to M2Server SUCCESS !!!\n", sess.SessionID)

		go forwardM2ToClient(sess, conn)
		return
	}
}

func forwardM2ToClient(sess *network.GateSession, conn net.Conn) {
	buf := make([]byte, 8192)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("!!! M2Server read error for session %d: %v !!!\n", sess.SessionID, err)
			}
			fmt.Printf("!!! M2Server connection lost for session %d !!!\n", sess.SessionID)
			sess.Close()
			return
		}

		fmt.Printf("!!! Session %d: received %d bytes from M2Server: %x !!!\n", sess.SessionID, n, buf[:n])

		data := buf[:n]
		if len(data) >= 6 && binary.LittleEndian.Uint32(data[0:4]) == protocol.RUNGATECODE {
			clientData := data[20:]
			if len(clientData) >= 14 {
				msgEncoded := encode6Bit(clientData[:14])
				bodyEncoded := encode6Bit(clientData[14:])
				packet := fmt.Sprintf("#1%s%s!", msgEncoded, bodyEncoded)
				fmt.Printf("!!! Session %d: sending encoded response: %s !!!\n", sess.SessionID, packet)
				sess.Send([]byte(packet))
				continue
			}
		}

		encoded := encode6Bit(data)
		packet := fmt.Sprintf("#1%s!", encoded)
		fmt.Printf("!!! Session %d: forwarding to client: %s !!!\n", sess.SessionID, packet)
		sess.Send([]byte(packet))
	}
}

func encode6Bit(data []byte) string {
	result := make([]byte, 0, len(data)*2)
	var buffer, bits int

	for i := 0; i < len(data); i++ {
		buffer = (buffer << 8) | int(data[i])
		bits += 8

		for bits >= 6 {
			bits -= 6
			result = append(result, byte((buffer>>bits)&0x3F+0x3C))
		}
	}

	if bits > 0 {
		result = append(result, byte((buffer<<(6-bits))&0x3F+0x3C))
	}

	return string(result)
}

func onDisconnect(sess *network.GateSession) {
	fmt.Printf("!!! Player disconnected: %s session=%d !!!\n", sess.Addr, sess.SessionID)
	logger.Info("Player disconnected from RunGate",
		zap.String("addr", sess.Addr),
		zap.Int32("session", sess.SessionID),
	)

	sessionsMutex.Lock()
	if cs, ok := clientSessions[sess.SessionID]; ok {
		if cs.m2Conn != nil {
			cs.m2Conn.Close()
		}
		delete(clientSessions, sess.SessionID)
	}
	sessionsMutex.Unlock()
}

func onRunMessage(sess *network.GateSession, data []byte) {
	fmt.Printf("!!! onRunMessage: session=%d, data_len=%d, data=%x !!!\n", sess.SessionID, len(data), data)
	dataStr := string(data)

	if len(dataStr) > 0 && dataStr[0] == '#' {
		if !strings.HasSuffix(dataStr, "!") {
			fmt.Printf("!!! Session %d: message doesn't end with ! !!!\n", sess.SessionID)
			return
		}

		sessionCode := int(dataStr[1] - '0')
		if sessionCode != 1 {
			fmt.Printf("!!! Session %d: wrong session code %d !!!\n", sess.SessionID, sessionCode)
			return
		}

		idx := strings.Index(dataStr, "!")
		if idx < 2 {
			fmt.Printf("!!! Session %d: invalid message format !!!\n", sess.SessionID)
			return
		}

		encoded := dataStr[2:idx]
		decoded := decode6Bit(encoded)

		fmt.Printf("!!! Client message: session=%d, encoded_len=%d, decoded_len=%d, decoded=%x !!!\n",
			sess.SessionID, len(encoded), len(decoded), decoded)

		if len(decoded) < 14 {
			fmt.Printf("!!! Session %d: decoded data too short !!!\n", sess.SessionID)
			return
		}

		ident := uint16(decoded[0]) | (uint16(decoded[1]) << 8)
		fmt.Printf("!!! Session %d: ident=%d !!!\n", sess.SessionID, ident)

		sessionsMutex.RLock()
		cs, ok := clientSessions[sess.SessionID]
		sessionsMutex.RUnlock()

		if !ok || cs == nil {
			fmt.Printf("!!! Session %d: no client session found !!!\n", sess.SessionID)
			return
		}

		if cs.m2Conn == nil {
			fmt.Printf("!!! Session %d: M2Server not connected, queuing message !!!\n", sess.SessionID)
			return
		}

		msgHeader := protocol.TMsgHeader{
			Code:          protocol.RUNGATECODE,
			Socket:        int32(sess.SessionID),
			GSocketIdx:    0,
			Ident:         protocol.GM_DATA,
			UserListIndex: 0,
			Length:        int32(len(decoded)),
		}

		headerData := msgHeader.Pack()
		packet := make([]byte, len(headerData)+len(decoded))
		copy(packet, headerData)
		copy(packet[len(headerData):], decoded)

		fmt.Printf("!!! Sending to M2Server: len=%d, header=%x, data=%x !!!\n", len(packet), headerData, decoded)

		_, err := cs.m2Conn.Write(packet)
		if err != nil {
			fmt.Printf("!!! Session %d: Failed to forward to M2Server: %v !!!\n", sess.SessionID, err)
			cs.m2Conn.Close()
			cs.m2Conn = nil
			go connectM2ForClient(sess, cs)
			return
		}

		fmt.Printf("!!! Session %d: forwarded %d bytes to M2Server !!!\n", sess.SessionID, len(packet))
		return
	}

	fmt.Printf("!!! Session %d: non-# message, len=%d !!!\n", sess.SessionID, len(data))
}

func decode6Bit(s string) []byte {
	result := make([]byte, 0, len(s)*6/8)
	var buffer, bits int

	for i := 0; i < len(s); i++ {
		ch := int(s[i])
		if ch < 0x3C {
			continue
		}
		value := ch - 0x3C
		buffer = (buffer << 6) | value
		bits += 6

		if bits >= 8 {
			bits -= 8
			result = append(result, byte(buffer>>bits))
			buffer &= (1 << bits) - 1
		}
	}
	return result
}
