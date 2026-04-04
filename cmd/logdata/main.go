package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/mir2go/mir2/pkg/config"
)

type LogEntry struct {
	LogType   int
	Time      time.Time
	MapName   string
	X, Y      int
	ActorName string
	ItemName  string
	Gold      int
	Exp       int
	Message   string
}

type LogDataServer struct {
	logger     *zap.Logger
	cfg        *config.ServerConfig
	LogQueue   chan string
	LogFile    *os.File
	WriteMutex sync.Mutex
	Running    bool
}

func main() {
	flag.Parse()

	zapLogger, _ := zap.NewProduction()
	logger := zapLogger
	defer logger.Sync()

	logger.Info("Starting LogDataServer...")

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		logger.Warn("Failed to load config, using default", zap.Error(err))
		cfg = config.GetDefaultConfig()
	}

	server := NewLogDataServer(logger, cfg)
	if err := server.Start(); err != nil {
		logger.Error("Failed to start server", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("LogDataServer started successfully")

	waitForSignal()
	logger.Info("LogDataServer shutting down...")
	server.Stop()
}

func NewLogDataServer(logger *zap.Logger, cfg *config.ServerConfig) *LogDataServer {
	return &LogDataServer{
		logger:   logger,
		cfg:      cfg,
		LogQueue: make(chan string, 10000),
		Running:  true,
	}
}

func (s *LogDataServer) Start() error {
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	dateStr := time.Now().Format("2006-01-02")
	logFileName := fmt.Sprintf("%s/%s.log", logDir, dateStr)
	logFile, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	s.LogFile = logFile

	go s.writeLogLoop()

	s.logger.Info("LogDataServer started", zap.String("log_file", logFileName))
	return nil
}

func (s *LogDataServer) Stop() {
	s.Running = false
	if s.LogFile != nil {
		s.LogFile.Close()
	}
}

func (s *LogDataServer) writeLogLoop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for s.Running {
		select {
		case logMsg := <-s.LogQueue:
			s.writeLog(logMsg)
		case <-ticker.C:
			s.flushLogs()
		}
	}
}

func (s *LogDataServer) writeLog(msg string) {
	s.WriteMutex.Lock()
	defer s.WriteMutex.Unlock()

	if s.LogFile != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logLine := fmt.Sprintf("[%s] %s\n", timestamp, msg)
		s.LogFile.WriteString(logLine)
	}
}

func (s *LogDataServer) flushLogs() {
	s.WriteMutex.Lock()
	defer s.WriteMutex.Unlock()

	if s.LogFile != nil {
		s.LogFile.Sync()
	}
}

func (s *LogDataServer) AddLog(logType int, mapName string, x, y int, actorName, itemName string, gold, exp int, message string) {
	entry := LogEntry{
		LogType:   logType,
		Time:      time.Now(),
		MapName:   mapName,
		X:         x,
		Y:         y,
		ActorName: actorName,
		ItemName:  itemName,
		Gold:      gold,
		Exp:       exp,
		Message:   message,
	}

	logMsg := fmt.Sprintf("Type:%d Map:%s X:%d Y:%d Actor:%s Item:%s Gold:%d Exp:%d Msg:%s",
		entry.LogType, entry.MapName, entry.X, entry.Y,
		entry.ActorName, entry.ItemName, entry.Gold, entry.Exp, entry.Message)

	select {
	case s.LogQueue <- logMsg:
	default:
		s.logger.Warn("Log queue full, dropping message")
	}
}

const (
	LOG_TYPE_NONE     = 0
	LOG_TYPE_GOLD     = 1
	LOG_TYPE_ITEM     = 2
	LOG_TYPE_MAP      = 3
	LOG_TYPE_MAKE     = 4
	LOG_TYPE_GUILD    = 5
	LOG_TYPE_MAIL     = 6
	LOG_TYPE_TRADE    = 7
	LOG_TYPE_DEALGOLD = 8
	LOG_TYPE_DEALITEM = 9
	LOG_TYPE_GAMEGOLD = 10
	LOG_TYPE_STORAGE  = 11
	LOG_TYPE_MISSION  = 12
)

func waitForSignal() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
}
