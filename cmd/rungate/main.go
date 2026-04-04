package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/mir2go/mir2/pkg/config"
	"github.com/mir2go/mir2/pkg/network"
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

	go connectM2ForClient(sess, cs)
}

func connectM2ForClient(sess *network.GateSession, cs *ClientSession) {
	for {
		fmt.Printf("!!! Connecting session %d to M2Server at %s !!!\n", sess.SessionID, m2Addr)
		conn, err := net.DialTimeout("tcp", m2Addr, 5*time.Second)
		if err != nil {
			fmt.Printf("!!! M2Server connection failed for session %d: %v, retrying !!!\n", sess.SessionID, err)
			time.Sleep(3 * time.Second)
			continue
		}

		cs.m2Conn = conn
		fmt.Printf("!!! Session %d connected to M2Server !!!\n", sess.SessionID)

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
		sess.Send(buf[:n])
	}
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
	if len(data) < 14 {
		return
	}

	ident := uint16(data[0]) | (uint16(data[1]) << 8)

	fmt.Printf("!!! Client message: session=%d, ident=%d, data_len=%d !!!\n",
		sess.SessionID, ident, len(data))

	sessionsMutex.RLock()
	cs, ok := clientSessions[sess.SessionID]
	sessionsMutex.RUnlock()

	if !ok || cs.m2Conn == nil {
		fmt.Printf("!!! Session %d: M2Server not connected !!!\n", sess.SessionID)
		return
	}

	_, err := cs.m2Conn.Write(data)
	if err != nil {
		fmt.Printf("!!! Session %d: Failed to forward to M2Server: %v !!!\n", sess.SessionID, err)
		cs.m2Conn.Close()
		cs.m2Conn = nil
		go connectM2ForClient(sess, cs)
		return
	}

	fmt.Printf("!!! Session %d: forwarded %d bytes to M2Server !!!\n", sess.SessionID, len(data))
}
