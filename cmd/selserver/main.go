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
	logger     *zap.Logger
	server     *network.GateServer
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

	logger.Info("Starting SelGate...")

	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		logger.Warn("Failed to load config, using default", zap.Error(err))
		cfg = config.GetDefaultConfig()
	}

	addr := fmt.Sprintf("%s:%d", cfg.SelGate.IP, cfg.SelGate.Port)
	server = network.NewGateServer(addr, logger)
	server.MaxSessions = cfg.SelGate.MaxConn
	server.OnConnect = onConnect
	server.OnDisconnect = onDisconnect
	server.OnMessage = onSelMessage

	if err := server.Start(); err != nil {
		logger.Error("Failed to start SelGate", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("SelGate started",
		zap.String("addr", addr),
		zap.Int("maxconn", cfg.SelGate.MaxConn),
	)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	logger.Info("Shutting down SelGate...")
	server.Stop()
}

func onConnect(sess *network.GateSession) {
	logger.Info("Client connected to SelGate",
		zap.String("addr", sess.Addr),
		zap.Int32("session", sess.SessionID),
	)
}

func onDisconnect(sess *network.GateSession) {
	logger.Info("Client disconnected from SelGate",
		zap.String("addr", sess.Addr),
		zap.Int32("session", sess.SessionID),
	)
}

func onSelMessage(sess *network.GateSession, data []byte) {
	if len(data) < 2 {
		return
	}

	ident := uint16(data[0]) | (uint16(data[1]) << 8)

	switch ident {
	case 100:
		logger.Debug("CM_QUERYCHR", zap.Int32("session", sess.SessionID))
		sendServerList(sess)
	case 101:
		logger.Debug("CM_NEWCHR", zap.Int32("session", sess.SessionID))
	case 102:
		logger.Debug("CM_DELCHR", zap.Int32("session", sess.SessionID))
	case 103:
		logger.Debug("CM_SELCHR", zap.Int32("session", sess.SessionID))
		handleSelectServer(sess, data)
	case 104:
		logger.Debug("CM_SELECTSERVER", zap.Int32("session", sess.SessionID))
	default:
		logger.Debug("Unknown SelGate message", zap.Uint16("ident", ident))
	}
}

func sendServerList(sess *network.GateSession) {
	cfg := config.GetDefaultConfig()

	logger.Info("Sending server list to client",
		zap.Int32("session", sess.SessionID),
		zap.Int("count", len(cfg.LoginSrv.ServerList)),
	)

	for _, srv := range cfg.LoginSrv.ServerList {
		logger.Debug("Server info",
			zap.String("name", srv.Name),
			zap.String("ip", srv.IP),
			zap.Int("port", srv.Port),
		)
	}
}

func handleSelectServer(sess *network.GateSession, data []byte) {
	if len(data) < 6 {
		return
	}

	serverIndex := int(data[2])

	logger.Info("Server selected",
		zap.Int32("session", sess.SessionID),
		zap.Int("index", serverIndex),
	)
}
