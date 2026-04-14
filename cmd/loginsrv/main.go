package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/mir2go/mir2/pkg/config"
	"github.com/mir2go/mir2/pkg/db"
	"github.com/mir2go/mir2/pkg/protocol"
)

var (
	logger     *zap.Logger
	database   db.Database
	configFile string
	cfg        *config.ServerConfig
	m2Conn     net.Conn
)

type UserInfo struct {
	sessionID   int
	account     string
	serverName  string
	selServer   bool
	sockIndex   string
	remoteIP    string
	conn        net.Conn
}

var (
	userList    = make(map[int]*UserInfo)
	userCount   int
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
	logger.Info("Listening on " + addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("Failed to listen", zap.Error(err))
		os.Exit(1)
	}

	logger.Info("LoginSrv started", zap.String("addr", addr))

	go connectToM2Server()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error("Accept error", zap.Error(err))
			continue
		}
		logger.Info("Client connected", zap.String("addr", conn.RemoteAddr().String()))
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()
	logger.Info("Client connected from: " + conn.RemoteAddr().String())

	buf := make([]byte, 8192)
	leftover := ""

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			logger.Info("Client disconnected:", zap.Error(err))
			break
		}

		data := leftover + string(buf[:n])
		logger.Debug("Received:", zap.String("data", data))

		for idx := strings.Index(data, "$"); idx >= 0; idx = strings.Index(data, "$") {
			line := data[:idx+1]
			data = data[idx+1:]
			line = strings.TrimSuffix(line, "$")

			if len(line) == 0 {
				continue
			}

			logger.Debug("Processing:", zap.String("msg", line))
			processMessage(conn, line)
		}
		leftover = data
	}

	for _, user := range userList {
		if user.conn == conn {
			delete(userList, user.sessionID)
			break
		}
	}
	conn.Close()
}

func processMessage(conn net.Conn, msg string) {
	if len(msg) < 2 || msg[0] != '%' {
		return
	}

	msg = msg[1:]

	switch msg[0] {
	case 'N':
		handleNewConnection(conn, msg[1:])
	case 'C':
		handleDisconnect(conn, msg[1:])
	case 'D':
		handleClientMessage(conn, msg[1:])
	default:
		if msg[0] >= '0' && msg[0] <= '9' {
			handleClientMessage(conn, msg)
		} else {
			logger.Debug("Unknown message type", zap.String("msg", msg))
		}
	}
}

func handleNewConnection(conn net.Conn, data string) {
	parts := strings.SplitN(data, "/", 3)
	if len(parts) < 2 {
		return
	}

	sessionID := userCount
	userCount++

	remoteIP := parts[0]
	if len(parts) > 1 {
		remoteIP = parts[1]
	}

	user := &UserInfo{
		sessionID:  sessionID,
		sockIndex:  fmt.Sprintf("%d", sessionID),
		remoteIP:   remoteIP,
		conn:       conn,
		selServer:  false,
	}
	userList[sessionID] = user

	logger.Info("New session", zap.Int("id", sessionID), zap.String("ip", remoteIP))

	sendServerName(conn, sessionID)
}

func handleDisconnect(conn net.Conn, data string) {
	data = strings.TrimSuffix(data, "$")
	sessionID := 0
	fmt.Sscanf(data, "%d", &sessionID)

	if user, ok := userList[sessionID]; ok {
		logger.Info("User disconnected", zap.Int("id", sessionID), zap.String("account", user.account))
		delete(userList, sessionID)
	}
}

func handleClientMessage(conn net.Conn, data string) {
	slashIdx := strings.Index(data, "/")
	if slashIdx < 0 {
		return
	}

	sessionID := 0
	fmt.Sscanf(data[:slashIdx], "%d", &sessionID)

	packet := data[slashIdx+1:]
	logger.Debug("Client packet", zap.Int("session", sessionID), zap.String("packet", packet))

	user, ok := userList[sessionID]
	if !ok {
		logger.Warn("Session not found", zap.Int("session", sessionID))
		return
	}

	if strings.HasPrefix(packet, "#") && strings.HasSuffix(packet, "!") {
		packet = packet[1 : len(packet)-1]
	}

	processClientMessage(conn, user, packet)
}

func processClientMessage(conn net.Conn, user *UserInfo, packet string) {
	if len(packet) < 1 {
		return
	}

	sessionIDChar := packet[0]
	encoded := packet[1:]

	raw := decode6Bit(encoded)

	if len(raw) < 12 {
		logger.Debug("Decoded data too short", zap.Int("len", len(raw)))
		return
	}

	ident := binary.LittleEndian.Uint16(raw[4:6])
	var body string
	if len(raw) > 12 {
		body = string(raw[12:])
	}

	logger.Info("Message", zap.String("sessionChar", string(sessionIDChar)), zap.Uint16("ident", ident), zap.String("body", body))

	switch ident {
	case 107, 100:
		logger.Info("CM_QUERYSERVERNAME")
		sendServerName(conn, user.sessionID)
	case 104:
		logger.Info("CM_SELECTSERVER", zap.String("server", body))
		user.selServer = true
		user.serverName = body
		handleSelectServer(conn, user, body)
	case 2001:
		logger.Info("CM_IDPASSWORD", zap.String("account", body))
		handleLogin(conn, user, body)
	default:
		logger.Debug("Unknown ident", zap.Uint16("ident", ident))
	}
}

func sendServerName(conn net.Conn, sessionID int) {
	serverCount := len(cfg.LoginSrv.ServerList)
	if serverCount == 0 {
		serverCount = 1
	}

	var serverInfo string
	for i, srv := range cfg.LoginSrv.ServerList {
		if i > 0 {
			serverInfo += "/"
		}
		serverInfo += srv.Name + "/0"
	}

	if serverInfo == "" {
		serverInfo = "Server1/0"
	}

	msg := make([]byte, 12)
	binary.LittleEndian.PutUint32(msg[0:4], 0)
	binary.LittleEndian.PutUint16(msg[4:6], 537)
	binary.LittleEndian.PutUint16(msg[6:8], 0)
	binary.LittleEndian.PutUint16(msg[8:10], 0)
	binary.LittleEndian.PutUint16(msg[10:12], uint16(serverCount))

	msgEncoded := encode6Bit(msg)
	bodyEncoded := encodeString(serverInfo)

	packet := fmt.Sprintf("#%d%s%s!", sessionID, msgEncoded, bodyEncoded)
	sendToClient(conn, sessionID, packet)

	logger.Info("Sent SM_SERVERNAME", zap.String("servers", serverInfo))
}

func handleSelectServer(conn net.Conn, user *UserInfo, serverName string) {
	var srv config.ServerInfo
	found := false
	for _, s := range cfg.LoginSrv.ServerList {
		if s.Name == serverName {
			srv = s
			found = true
			break
		}
	}

	if !found {
		srv = cfg.LoginSrv.ServerList[0]
	}

	addr := fmt.Sprintf("%s:%d", srv.IP, srv.Port)

	msg := make([]byte, 12)
	binary.LittleEndian.PutUint16(msg[4:6], 530)

	response := make([]byte, 4+len(addr))
	binary.LittleEndian.PutUint32(response[0:4], 0)
	copy(response[4:], []byte(addr))

	data := make([]byte, len(msg)+len(response))
	copy(data, msg)
	copy(data[len(msg):], response)

	encoded := encode6Bit(data)
	packet := fmt.Sprintf("#%d%s!", user.sessionID, encoded)
	sendToClient(conn, user.sessionID, packet)

	logger.Info("Sent SM_SELECTSERVER_OK", zap.String("addr", addr))
}

func handleLogin(conn net.Conn, user *UserInfo, data string) {
	parts := strings.Split(data, "/")
	if len(parts) < 2 {
		return
	}

	account := parts[0]
	password := parts[1]

	logger.Info("Login attempt", zap.String("account", account))

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

	user.account = account

	sendLoginResult(conn, user.sessionID, result)

	if result == 0x00 && user.serverName != "" {
		go sendSessionToM2Server(account, user.sessionID, user.serverName)
	}
}

func sendLoginResult(conn net.Conn, sessionID int, result byte) {
	count := len(cfg.LoginSrv.ServerList)
	if count == 0 {
		count = 1
	}

	msg := make([]byte, 12)
	binary.LittleEndian.PutUint16(msg[4:6], 529)
	binary.LittleEndian.PutUint16(msg[10:12], uint16(count))

	encoded := encode6Bit(msg)
	packet := fmt.Sprintf("#%d%s!", sessionID, encoded)
	sendToClient(conn, sessionID, packet)

	logger.Info("Sent login result", zap.String("result", fmt.Sprintf("%d", result)))
}

func connectToM2Server() {
	m2Addr := fmt.Sprintf("127.0.0.1:%d", cfg.M2Server.Port)
	for {
		conn, err := net.DialTimeout("tcp", m2Addr, 5*time.Second)
		if err != nil {
			logger.Warn("Failed to connect to M2Server, retrying in 5s", zap.Error(err))
			time.Sleep(5 * time.Second)
			continue
		}
		m2Conn = conn
		logger.Info("Connected to M2Server")
		go handleM2ServerMessages()
		return
	}
}

func handleM2ServerMessages() {
	buf := make([]byte, 8192)
	for {
		if m2Conn == nil {
			return
		}
		m2Conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := m2Conn.Read(buf)
		if err != nil {
			logger.Warn("M2Server connection lost, reconnecting", zap.Error(err))
			m2Conn = nil
			go connectToM2Server()
			return
		}
		data := buf[:n]
		logger.Debug("Received from M2Server", zap.ByteString("data", data))
	}
}

func sendSessionToM2Server(account string, sessionID int, serverName string) {
	for i := 0; i < 10; i++ {
		if m2Conn != nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	if m2Conn == nil {
		logger.Warn("M2Server not connected, cannot send session")
		return
	}

	msg := fmt.Sprintf("%s/%d/%s", account, sessionID, serverName)
	data := make([]byte, 6+len(msg))
	binary.LittleEndian.PutUint32(data[0:4], 0xAA55AA55)
	binary.LittleEndian.PutUint16(data[4:6], uint16(protocol.SS_OPENSESSION))
	copy(data[6:], []byte(msg))

	logger.Info("Sending SS_OPENSESSION to M2Server", zap.ByteString("data", data), zap.String("msg", msg))
	
	_, err := m2Conn.Write(data)
	if err != nil {
		logger.Error("Failed to send session to M2Server", zap.Error(err))
	} else {
		logger.Info("Sent SS_OPENSESSION to M2Server", zap.String("account", account), zap.Int("sessionID", sessionID))
	}
}

func sendToClient(conn net.Conn, sessionID int, packet string) {
	msg := fmt.Sprintf("%%%d/%s$", sessionID, packet)
	conn.Write([]byte(msg))
	logger.Debug("Sent to client", zap.String("msg", msg))
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

func encodeString(s string) string {
	return encode6Bit([]byte(s))
}

func decodeString(s string) string {
	return string(decode6Bit(s))
}
