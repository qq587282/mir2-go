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
	cfg       *config.ServerConfig
)

func init() {
	flag.StringVar(&configFile, "config", "config.yaml", "config file path")
}

func main() {
	flag.Parse()
	
	zapLogger, _ := zap.NewProduction()
	logger = zapLogger
	defer logger.Sync()
	
	logger.Info("Starting LoginSrv...")
	
	var err error
	cfg, err = config.LoadConfig(configFile)
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
	
	database, err = db.NewDatabase(dbCfg)
	if err != nil {
		logger.Warn("Failed to connect database, using memory storage", zap.Error(err))
	}
	
	addr := fmt.Sprintf("%s:%d", cfg.LoginSrv.IP, cfg.LoginSrv.Port)
	server = network.NewGateServer(addr, logger)
	server.MaxSessions = 10000
	server.OnConnect = onConnect
	server.OnDisconnect = onDisconnect
	server.OnMessage = onLoginSrvMessage
	
	if err := server.Start(); err != nil {
		logger.Error("Failed to start LoginSrv", zap.Error(err))
		os.Exit(1)
	}
	
	logger.Info("LoginSrv started",
		zap.String("addr", addr),
		zap.Strings("servers", getServerNames(cfg.LoginSrv.ServerList)),
	)
	
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	logger.Info("Shutting down LoginSrv...")
	if database != nil {
		database.Close()
	}
	server.Stop()
}

func getServerNames(list []config.ServerInfo) []string {
	names := make([]string, len(list))
	for i, s := range list {
		names[i] = s.Name
	}
	return names
}

func onConnect(sess *network.GateSession) {
	logger.Info("Client connected to LoginSrv",
		zap.String("addr", sess.Addr),
		zap.Int32("session", sess.SessionID),
	)
}

func onDisconnect(sess *network.GateSession) {
	logger.Info("Client disconnected from LoginSrv",
		zap.String("addr", sess.Addr),
		zap.Int32("session", sess.SessionID),
	)
}

func onLoginSrvMessage(sess *network.GateSession, data []byte) {
	if len(data) < 14 {
		return
	}
	
	ident := uint16(data[0]) | (uint16(data[1]) << 8)
	
	switch ident {
	case 2001:
		logger.Debug("CM_IDPASSWORD", zap.Int32("session", sess.SessionID))
		handleLogin(sess, data)
	case 2002:
		logger.Debug("CM_ADDNEWUSER", zap.Int32("session", sess.SessionID))
		handleNewAccount(sess, data)
	case 2003:
		logger.Debug("CM_CHANGEPASSWORD", zap.Int32("session", sess.SessionID))
	case 104:
		logger.Debug("CM_SELECTSERVER", zap.Int32("session", sess.SessionID))
		handleSelectServer(sess, data)
	default:
		logger.Debug("Unknown LoginSrv message", zap.Uint16("ident", ident))
	}
}

func handleLogin(sess *network.GateSession, data []byte) {
	body := data[14:]
	
	if len(body) < 60 {
		return
	}
	
	account := trimString(body[:30])
	password := trimString(body[30:60])
	
	logger.Info("Login attempt",
		zap.Int32("session", sess.SessionID),
		zap.String("account", account),
	)
	
	var result byte = 0x00
	
	if database != nil {
		acc, err := database.GetAccount(account)
		if err != nil {
			logger.Error("Failed to get account", zap.Error(err))
			result = 0x01
		} else if acc == nil {
			result = 0x02
		} else if acc.LoginPW != password {
			result = 0x03
		} else {
			result = 0x00
		}
	}
	
	sendLoginResult(sess, result)
}

func handleNewAccount(sess *network.GateSession, data []byte) {
	body := data[14:]
	
	if len(body) < 60 {
		return
	}
	
	account := trimString(body[:30])
	password := trimString(body[30:60])
	
	logger.Info("New account attempt",
		zap.Int32("session", sess.SessionID),
		zap.String("account", account),
	)
	
	var result byte = 0x00
	
	if database != nil {
		acc, err := database.GetAccount(account)
		if err != nil {
			logger.Error("Failed to get account", zap.Error(err))
			result = 0x01
		} else if acc != nil {
			result = 0x04
		} else {
			err = database.CreateAccount(account, password, "")
			if err != nil {
				logger.Error("Failed to create account", zap.Error(err))
				result = 0x01
			}
		}
	}
	
	sendLoginResult(sess, result)
}

func handleSelectServer(sess *network.GateSession, data []byte) {
	if len(data) < 16 {
		return
	}
	
	serverIndex := int(data[14])
	
	logger.Info("Server selected",
		zap.Int32("session", sess.SessionID),
		zap.Int("index", serverIndex),
	)
	
	sendServerInfo(sess, serverIndex)
}

func sendLoginResult(sess *network.GateSession, result byte) {
	response := make([]byte, 14)
	response[0] = 0x01
	response[1] = 0x00
	response[2] = result
	
	sess.Send(network.EncodePacket(response))
}

func sendServerInfo(sess *network.GateSession, serverIndex int) {
	if serverIndex >= len(cfg.LoginSrv.ServerList) {
		serverIndex = 0
	}
	
	serverInfo := cfg.LoginSrv.ServerList[serverIndex]
	
	msg := fmt.Sprintf("%s|%s|%d|%s|%s",
		serverInfo.Name,
		serverInfo.IP,
		serverInfo.Port,
		serverInfo.WebURL,
		serverInfo.Tag,
	)
	
	response := make([]byte, 14+len(msg))
	response[0] = 0x10
	response[1] = 0x00
	copy(response[14:], msg)
	
	sess.Send(network.EncodePacket(response))
}

func trimString(data []byte) string {
	end := 0
	for i, b := range data {
		if b == 0 {
			end = i
			break
		}
		end = i + 1
	}
	return string(data[:end])
}
