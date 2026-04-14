package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/mir2go/mir2/pkg/config"
)

var configFile string

type Session struct {
	conn       net.Conn
	index      int
	remoteIP   string
	reviceMsg  string
}

var (
	cfg          *config.ServerConfig
	sessions     = make(map[int]*Session)
	sessionIndex int
	logFile      *os.File
	writeLog     func(string)
	loginSrvConn net.Conn
)

func init() {
	flag.StringVar(&configFile, "config", "config.yaml", "config file path")
}

func main() {
	flag.Parse()
	os.Chdir("D:\\code\\mir2-go")

	logFile, _ = os.Create("logingate.log")
	defer logFile.Close()

	writeLog = func(msg string) {
		fmt.Println(msg)
		logFile.WriteString(msg + "\n")
		logFile.Sync()
	}

	writeLog("=== LoginGate Starting ===")

	var err error
	cfg, err = config.LoadConfig(configFile)
	if err != nil || cfg == nil {
		cfg = config.GetDefaultConfig()
	}

	loginSrvAddr := fmt.Sprintf("127.0.0.1:%d", cfg.LoginSrv.Port)
	writeLog("Connecting to LoginSrv: " + loginSrvAddr)
	loginSrvConn, err = net.Dial("tcp", loginSrvAddr)
	if err != nil {
		writeLog("Failed to connect to LoginSrv: " + err.Error())
		os.Exit(1)
	}
	writeLog("Connected to LoginSrv")
	go readFromLoginSrv()

	addr := fmt.Sprintf("%s:%d", cfg.LoginGate.IP, cfg.LoginGate.Port)
	writeLog("Listening on " + addr)

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		writeLog("ERROR: " + err.Error())
		os.Exit(1)
	}

	writeLog("=== LoginGate Started ===")

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		writeLog("New connection: " + conn.RemoteAddr().String())
		go handleClient(conn)
	}
}

func readFromLoginSrv() {
	buf := make([]byte, 8192)
	for {
		if loginSrvConn == nil {
			break
		}
		loginSrvConn.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := loginSrvConn.Read(buf)
		if err != nil {
			writeLog("LoginSrv read error: " + err.Error())
			break
		}

		data := string(buf[:n])
		writeLog("From LoginSrv RAW: " + string(buf[:n]))
		writeLog("From LoginSrv: " + data)

		parseAndForwardToClient(data)
	}
}

func parseAndForwardToClient(data string) {
  for len(data) > 0 {
    idx := strings.Index(data, "!")
    if idx < 0 {
      break
    }

    packet := data[:idx+1]
    data = data[idx+1:]

    if len(packet) < 2 {
      continue
    }

    if packet[0] == '%' {
      slashIdx := strings.Index(packet[1:], "/")
      if slashIdx < 0 {
        continue
      }

      sessionIdx, _ := strconv.Atoi(packet[1:slashIdx])
      msg := packet[slashIdx+1:]

      session := sessions[sessionIdx]
      if session != nil && session.conn != nil {
        session.conn.Write([]byte(msg))
        writeLog(fmt.Sprintf("Forwarded to client %d: %s", sessionIdx, msg))
      }
    } else if packet[0] == '#' {
      slashIdx := strings.Index(packet[1:], "/")
      if slashIdx < 0 {
        continue
      }

      sessionIdx, _ := strconv.Atoi(packet[1 : slashIdx+1])
      msg := packet[slashIdx+2 : len(packet)-1]

      session := sessions[sessionIdx]
      if session != nil && session.conn != nil {
        session.conn.Write([]byte(packet))
        writeLog(fmt.Sprintf("Forwarded to client %d: %s", sessionIdx, msg))
      }
    }
  }
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	sessionIdx := sessionIndex
	sessionIndex++

	session := &Session{
		conn:     conn,
		index:    sessionIdx,
		remoteIP: conn.RemoteAddr().String(),
	}
	sessions[sessionIdx] = session

	writeLog(fmt.Sprintf("Client %d connected: %s", sessionIdx, conn.RemoteAddr().String()))

	defer func() {
		delete(sessions, sessionIdx)
	}()

	notifyLoginSrvConnect(sessionIdx, session.remoteIP)

	buf := make([]byte, 4096)
	leftover := ""

	for {
		conn.SetReadDeadline(time.Now().Add(30 * time.Second))
		n, err := conn.Read(buf)
		if err != nil {
			writeLog(fmt.Sprintf("Client %d disconnected: %v", sessionIdx, err))
			break
		}

		data := leftover + string(buf[:n])
		writeLog(fmt.Sprintf("Client %d raw: %s", sessionIdx, string(buf[:n])))

		for idx := strings.Index(data, "!"); idx >= 0; idx = strings.Index(data, "!") {
			line := data[:idx+1]
			data = data[idx+1:]
			line = strings.TrimSuffix(line, "!")

			if len(line) == 0 || line[0] != '#' {
				continue
			}

			writeLog(fmt.Sprintf("Client %d: %s", sessionIdx, line))
			forwardToLoginSrv(sessionIdx, line)
		}
		leftover = data
	}

	notifyLoginSrvDisconnect(sessionIdx)
}

func notifyLoginSrvConnect(sessionIdx int, remoteIP string) {
	if loginSrvConn != nil {
		msg := fmt.Sprintf("%%N%d/%s/%s$", sessionIdx, remoteIP, remoteIP)
		loginSrvConn.Write([]byte(msg))
		writeLog(fmt.Sprintf("Notified LoginSrv: %s", msg))
	}
}

func notifyLoginSrvDisconnect(sessionIdx int) {
	if loginSrvConn != nil {
		msg := fmt.Sprintf("%%C%d$", sessionIdx)
		loginSrvConn.Write([]byte(msg))
		writeLog(fmt.Sprintf("Notified LoginSrv disconnect: %s", msg))
	}
}

func forwardToLoginSrv(sessionIdx int, packet string) {
	if loginSrvConn != nil {
		msg := fmt.Sprintf("%%D%d/%s$", sessionIdx, packet)
		writeLog("Forwarding to LoginSrv RAW: " + msg)
		loginSrvConn.Write([]byte(msg))
		writeLog(fmt.Sprintf("Forwarded to LoginSrv: %s", msg))
	}
}

func sendServerName(conn net.Conn, code int, writeLog func(string)) {
	cfg := config.GetDefaultConfig()
	serverCount := len(cfg.LoginSrv.ServerList)
	if serverCount == 0 {
		serverCount = 1
	}

	var serverInfo string
	for i, srv := range cfg.LoginSrv.ServerList {
		if i > 0 {
			serverInfo += "/"
		}
		serverInfo += srv.Name + "/" + "0"
	}
	
	if serverInfo == "" {
		serverInfo = "Server1/0"
	}
	serverCount = len(cfg.LoginSrv.ServerList)
	if serverCount == 0 {
		serverCount = 1
	}

	writeLog("Sending SM_SERVERNAME (ident=537): " + serverInfo)

	msg := make([]byte, 12)
	binary.LittleEndian.PutUint32(msg[0:4], 0)
	binary.LittleEndian.PutUint16(msg[4:6], 537)
	binary.LittleEndian.PutUint16(msg[6:8], 0)
	binary.LittleEndian.PutUint16(msg[8:10], 0)
	binary.LittleEndian.PutUint16(msg[10:12], uint16(serverCount))

	msgEncoded := encode6Bit(msg)
	bodyEncoded := encodeString(serverInfo)

	packet := fmt.Sprintf("#%d%s%s!", code, msgEncoded, bodyEncoded)

	writeLog(fmt.Sprintf("Packet (%d bytes): %s", len(packet), packet))

	n, err := conn.Write([]byte(packet))
	if err != nil {
		writeLog(fmt.Sprintf("Write error: %v", err))
	} else {
		writeLog(fmt.Sprintf("Wrote %d bytes", n))
	}
}

func handleSelectServer(conn net.Conn, code int, writeLog func(string)) {
	cfg := config.GetDefaultConfig()
	if len(cfg.LoginSrv.ServerList) == 0 {
		writeLog("No servers configured")
		return
	}

	srv := cfg.LoginSrv.ServerList[0]
	addr := fmt.Sprintf("%s:%d", srv.IP, srv.Port)

	writeLog("Selected server: " + addr)

	msg := make([]byte, 12)
	binary.LittleEndian.PutUint16(msg[4:6], 530) // SM_SELECTSERVER_OK

	response := make([]byte, 4+len(addr))
	binary.LittleEndian.PutUint32(response[0:4], 0)
	copy(response[4:], []byte(addr))

	data := make([]byte, len(msg)+len(response))
	copy(data, msg)
	copy(data[len(msg):], response)

	encoded := encode6Bit(data)
	packet := fmt.Sprintf("#%d%s!", code, encoded)
	conn.Write([]byte(packet))

	writeLog("Sent SM_SELECTSERVER_OK (530)")
}

func sendLoginOK(conn net.Conn, code int, writeLog func(string)) {
	cfg := config.GetDefaultConfig()
	count := len(cfg.LoginSrv.ServerList)

	msg := make([]byte, 12)
	binary.LittleEndian.PutUint16(msg[4:6], 529) // SM_PASSOK_SELECTSERVER
	binary.LittleEndian.PutUint16(msg[10:12], uint16(count))

	encoded := encode6Bit(msg)
	packet := fmt.Sprintf("#%d%s!", code, encoded)
	conn.Write([]byte(packet))

	writeLog(fmt.Sprintf("Sent login OK, servers=%d", count))
}

func sendLoginSuccess(conn net.Conn, code int, writeLog func(string)) {
	msg := make([]byte, 12)
	binary.LittleEndian.PutUint16(msg[4:6], 119) // SM_LOGIN_SUCCESS
	binary.LittleEndian.PutUint32(msg[0:4], 1)  // session ID

	encoded := encode6Bit(msg)
	packet := fmt.Sprintf("#%d%s!", code, encoded)
	conn.Write([]byte(packet))

	writeLog("Sent SM_LOGIN_SUCCESS (119)")
}

func sendSelectServerOK(conn net.Conn, code int, writeLog func(string)) {
	cfg := config.GetDefaultConfig()
	srv := cfg.LoginSrv.ServerList[0]
	addr := fmt.Sprintf("%s:%d", srv.IP, srv.Port)

	msg := make([]byte, 12)
	binary.LittleEndian.PutUint16(msg[4:6], 530) // SM_SELECTSERVER_OK

	response := make([]byte, 4+len(addr))
	binary.LittleEndian.PutUint32(response[0:4], 0)
	copy(response[4:], []byte(addr))

	data := make([]byte, len(msg)+len(response))
	copy(data, msg)
	copy(data[len(msg):], response)

	encoded := encode6Bit(data)
	packet := fmt.Sprintf("#%d%s!", code, encoded)
	conn.Write([]byte(packet))

	writeLog("Sent SM_SELECTSERVER_OK (530): " + addr)
}

func sendPassOkSelectServer(conn net.Conn, code int, writeLog func(string)) {
	cfg := config.GetDefaultConfig()
	srv := cfg.LoginSrv.ServerList[0]
	addr := fmt.Sprintf("%s:%d", srv.IP, srv.Port)
	serverCount := len(cfg.LoginSrv.ServerList)
	if serverCount == 0 {
		serverCount = 1
	}

	msg := make([]byte, 12)
	binary.LittleEndian.PutUint16(msg[4:6], 529) // SM_PASSOK_SELECTSERVER
	binary.LittleEndian.PutUint16(msg[10:12], uint16(serverCount))

	body := addr + "/0/0"
	bodyEncoded := encodeString(body)

	data := make([]byte, len(msg)+len(bodyEncoded))
	copy(data, msg)
	copy(data[len(msg):], bodyEncoded)

	encoded := encode6Bit(data)
	packet := fmt.Sprintf("#%d%s!", code, encoded)
	conn.Write([]byte(packet))

	writeLog("Sent SM_PASSOK_SELECTSERVER (529) with encoded body: " + body)
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
