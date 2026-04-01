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
	
	zapLogger, _ := zap.NewProduction()
	logger = zapLogger
	defer logger.Sync()
	
	logger.Info("Starting RunGate...")
	
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		logger.Warn("Failed to load config, using default", zap.Error(err))
		cfg = config.GetDefaultConfig()
	}
	
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
	
	logger.Info("RunGate started",
		zap.String("addr", addr),
		zap.Int("maxconn", cfg.RunGate.MaxConn),
	)
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	logger.Info("Shutting down RunGate...")
	server.Stop()
}

func onConnect(sess *network.GateSession) {
	logger.Info("Player connected to RunGate",
		zap.String("addr", sess.Addr),
		zap.Int32("session", sess.SessionID),
	)
}

func onDisconnect(sess *network.GateSession) {
	logger.Info("Player disconnected from RunGate",
		zap.String("addr", sess.Addr),
		zap.Int32("session", sess.SessionID),
	)
}

func onRunMessage(sess *network.GateSession, data []byte) {
	if len(data) < 14 {
		return
	}
	
	ident := uint16(data[0]) | (uint16(data[1]) << 8)
	param := uint16(data[2]) | (uint16(data[3]) << 8)
	
	logger.Debug("RunGate message",
		zap.Int32("session", sess.SessionID),
		zap.Uint16("ident", ident),
		zap.Uint16("param", param),
	)
	
	switch ident {
	case 3010:
		logger.Debug("CM_TURN")
	case 3011:
		logger.Debug("CM_WALK")
	case 3013:
		logger.Debug("CM_RUN")
	case 3014:
		logger.Debug("CM_HIT")
	case 3017:
		logger.Debug("CM_SPELL")
	case 3030:
		if len(data) > 14 {
			msg := string(data[14:])
			logger.Debug("CM_SAY", zap.String("msg", msg))
		}
	default:
		logger.Debug("RunGate other message", zap.Uint16("ident", ident))
	}
}
