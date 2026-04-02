package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/mir2go/mir2/pkg/config"
	"github.com/mir2go/mir2/pkg/db"
	"github.com/mir2go/mir2/pkg/network"
)

var (
	logger    *zap.Logger
	server    *network.GateServer
	database  db.Database
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
	
	logger.Info("Starting DBServer...")
	
	cfg, err := config.LoadConfig(configFile)
	if err != nil {
		logger.Warn("Failed to load config, using default", zap.Error(err))
		cfg = config.GetDefaultConfig()
	}
	
	dbCfg := db.DBConfig{
		Type:     cfg.DBServer.Type,
		Host:     cfg.DBServer.IP,
		Port:     cfg.DBServer.Port,
		User:     cfg.DBServer.User,
		Password: cfg.DBServer.Password,
		Database: cfg.DBServer.Database,
	}
	
	if cfg.DBServer.Enable && dbCfg.Type != "memory" {
		database, err = db.NewDatabase(dbCfg)
		if err != nil {
			logger.Error("Failed to connect database", zap.Error(err))
			logger.Info("Using memory storage instead")
		} else {
			defer database.Close()
		}
	} else {
		logger.Info("Using memory storage")
	}
	
	addr := fmt.Sprintf("%s:%d", cfg.ServerIP, 5400)
	server = network.NewGateServer(addr, logger)
	server.MaxSessions = 100
	server.OnConnect = onConnect
	server.OnDisconnect = onDisconnect
	server.OnMessage = onDBMessage
	
	if err := server.Start(); err != nil {
		logger.Error("Failed to start DBServer", zap.Error(err))
		os.Exit(1)
	}
	
	logger.Info("DBServer started",
		zap.String("addr", addr),
		zap.String("dbtype", cfg.DBServer.Type),
	)
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	logger.Info("Shutting down DBServer...")
	server.Stop()
}

func onConnect(sess *network.GateSession) {
	logger.Info("Server connected to DBServer",
		zap.String("addr", sess.Addr),
		zap.Int32("session", sess.SessionID),
	)
}

func onDisconnect(sess *network.GateSession) {
	logger.Info("Server disconnected from DBServer",
		zap.String("addr", sess.Addr),
		zap.Int32("session", sess.SessionID),
	)
}

func onDBMessage(sess *network.GateSession, data []byte) {
	if len(data) < 2 {
		return
	}
	
	ident := uint16(data[0]) | (uint16(data[1]) << 8)
	
	switch ident {
	case 1:
		handleSaveCharacter(sess, data)
	case 2:
		handleLoadCharacter(sess, data)
	case 3:
		handleSaveHero(sess, data)
	case 4:
		handleLoadHero(sess, data)
	default:
		logger.Debug("Unknown DBServer message", zap.Uint16("ident", ident))
	}
}

func handleSaveCharacter(sess *network.GateSession, data []byte) {
	logger.Debug("Save character request")
	
	response := make([]byte, 4)
	response[0] = 0x01
	response[1] = 0x00
	response[2] = 0x00
	response[3] = 0x00
	
	sess.Send(network.EncodePacket(response))
}

func handleLoadCharacter(sess *network.GateSession, data []byte) {
	logger.Debug("Load character request")
	
	response := make([]byte, 4)
	response[0] = 0x02
	response[1] = 0x00
	response[2] = 0x00
	response[3] = 0x00
	
	sess.Send(network.EncodePacket(response))
}

func handleSaveHero(sess *network.GateSession, data []byte) {
	logger.Debug("Save hero request")
	
	response := make([]byte, 4)
	response[0] = 0x03
	response[1] = 0x00
	response[2] = 0x00
	response[3] = 0x00
	
	sess.Send(network.EncodePacket(response))
}

func handleLoadHero(sess *network.GateSession, data []byte) {
	logger.Debug("Load hero request")
	
	response := make([]byte, 4)
	response[0] = 0x04
	response[1] = 0x00
	response[2] = 0x00
	response[3] = 0x00
	
	sess.Send(network.EncodePacket(response))
}
