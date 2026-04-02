package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/mir2go/mir2/pkg/config"
	"github.com/mir2go/mir2/pkg/network"
)

var (
	logger    *zap.Logger
	server    *network.GateServer
	configFile string
)

func init() {
	flag.StringVar(&configFile, "config", "config.yaml", "config file path")
}

func main() {
	flag.Parse()
	
	zapLogger, _ := zap.NewDevelopment()
	logger = zapLogger
	defer logger.Sync()
	
	logger.Info("Starting LoginGate...")
	
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		logger.Warn("Failed to load config, using default", zap.Error(err))
		cfg = config.GetDefaultConfig()
	}
	
	addr := fmt.Sprintf("%s:%d", cfg.LoginGate.IP, cfg.LoginGate.Port)
	server = network.NewGateServer(addr, logger)
	server.MaxSessions = cfg.LoginGate.MaxConn
	server.OnConnect = onConnect
	server.OnDisconnect = onDisconnect
	server.OnMessage = onLoginMessage
	
	if err := server.Start(); err != nil {
		logger.Error("Failed to start LoginGate", zap.Error(err))
		os.Exit(1)
	}
	
	logger.Info("LoginGate started",
		zap.String("addr", addr),
		zap.Int("maxconn", cfg.LoginGate.MaxConn),
	)
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	logger.Info("Shutting down LoginGate...")
	server.Stop()
}

func onConnect(sess *network.GateSession) {
	logger.Info("Client connected to LoginGate",
		zap.String("addr", sess.Addr),
		zap.Int32("session", sess.SessionID),
	)

	sendBuf := make([]byte, 4)
	sendBuf[0] = 0x00
	sendBuf[1] = 0x00
	sendBuf[2] = 0x00
	sendBuf[3] = 0x00
	sess.Send(network.EncodePacket(sendBuf))
}

func onDisconnect(sess *network.GateSession) {
	logger.Info("Client disconnected from LoginGate",
		zap.String("addr", sess.Addr),
		zap.Int32("session", sess.SessionID),
	)
}

func onLoginMessage(sess *network.GateSession, data []byte) {
	dataStr := fmt.Sprintf("%x", data)
	logger.Info(">>> RAW DATA received",
		zap.Int32("session", sess.SessionID),
		zap.String("raw", dataStr),
		zap.Int("len", len(data)),
	)
	
	if len(data) < 2 {
		logger.Info("Data too short", zap.Int("len", len(data)))
		return
	}

	ident := uint16(data[0]) | (uint16(data[1]) << 8)
	logger.Info(">>> ident", zap.Int32("session", sess.SessionID), zap.Uint16("ident", ident))
	
	switch ident {
	case 2000:
		logger.Info("CM_PROTOCOL", zap.Int32("session", sess.SessionID))
	case 2001:
		logger.Info("CM_IDPASSWORD", zap.Int32("session", sess.SessionID))
		handleLogin(sess, data)
	case 2002:
		logger.Info("CM_ADDNEWUSER", zap.Int32("session", sess.SessionID))
	case 2003:
		logger.Info("CM_CHANGEPASSWORD", zap.Int32("session", sess.SessionID))
	default:
		logger.Info("Unknown message", zap.Uint16("ident", ident))
	}
}

func handleLogin(sess *network.GateSession, data []byte) {
	body := data[2:]
	
	if len(body) < 60 {
		return
	}
	
	account := string(body[:30])
	
	logger.Info("Login attempt",
		zap.Int32("session", sess.SessionID),
		zap.String("account", account),
	)
	
	response := make([]byte, 4)
	response[0] = 0x53
	response[1] = 0x00
	response[2] = 0x00
	response[3] = 0x00
	
	sess.Send(network.EncodePacket(response))
}
