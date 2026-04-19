package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

const (
	LoginGateHost = "127.0.0.1"
	LoginGatePort = 7000

	LoginSrvHost = "127.0.0.1"
	LoginSrvPort = 15500

	M2ServerHost = "127.0.0.1"
	M2ServerPort = 16000

	RunGateHost = "127.0.0.1"
	RunGatePort = 7200
)

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════╗")
	fmt.Println("║         Mir2 统一测试客户端 v1.0                     ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	testLoginGate(reader)
	testLoginSrv(reader)
	testM2Server(reader)

	fmt.Println("\n╔══════════════════════════════════════════════════════╗")
	fmt.Println("║              所有测试完成                              ║")
	fmt.Println("╚══════════════════════════════════════════════════════╝")
}

func testLoginGate(reader *bufio.Reader) {
	fmt.Println("═══════════════════════════════════════════════════════")
	fmt.Println("  测试1: LoginGate (7000)")
	fmt.Println("═══════════════════════════════════════════════════════")

	fmt.Println("\n[1.1] 连接到 LoginGate...")
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", LoginGateHost, LoginGatePort), 5*time.Second)
	if err != nil {
		fmt.Printf("  ❌ 连接失败: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Println("  ✅ 已连接")

	fmt.Println("\n[1.2] 发送新连接消息...")
	sendNewConnection(conn)
	processResponses(conn, 3)
	waitEnter(reader)

	fmt.Println("\n[1.3] 查询服务器列表 (CM_QUERYSERVERNAME 107)...")
	sendMessage(conn, 107, "")
	processResponses(conn, 3)
	waitEnter(reader)

	fmt.Println("\n[1.4] 选择服务器 (CM_SELECTSERVER 104)...")
	sendSelectServer(conn, 0)
	processResponses(conn, 3)
	waitEnter(reader)

	fmt.Println("\n[1.5] 登录验证 (CM_IDPASSWORD 2001)...")
	sendLogin(conn, "testuser", "password")
	processResponses(conn, 5)

	fmt.Println("\n  ✅ LoginGate 测试完成")
	waitEnter(reader)
}

func testLoginSrv(reader *bufio.Reader) {
	fmt.Println("\n═══════════════════════════════════════════════════════")
	fmt.Println("  测试2: LoginSrv (15500)")
	fmt.Println("═══════════════════════════════════════════════════════")

	loginSrvSessionID = 0
	fmt.Println("\n[2.1] 连接到 LoginSrv...")
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", LoginSrvHost, LoginSrvPort), 5*time.Second)
	if err != nil {
		fmt.Printf("  ❌ 连接失败: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Println("  ✅ 已连接")

	fmt.Println("\n[2.2] 发送新连接消息...")
	sendNewConnectionLoginSrv(conn)
	processLoginSrvResponse(conn, 3)

	if loginSrvSessionID == 0 {
		loginSrvSessionID = 1
	}
	waitEnter(reader)

	fmt.Println("\n[2.3] 查询服务器列表 (CM_QUERYSERVERNAME 107)...")
	sendQueryServerLoginSrv(conn, loginSrvSessionID)
	processLoginSrvResponse(conn, 3)
	waitEnter(reader)

	fmt.Println("\n[2.4] 选择服务器 (CM_SELECTSERVER 104)...")
	sendSelectServerLoginSrv(conn, loginSrvSessionID, 0)
	processLoginSrvResponse(conn, 3)
	waitEnter(reader)

	fmt.Println("\n[2.5] 登录验证 (CM_IDPASSWORD 2001)...")
	sendLoginLoginSrv(conn, loginSrvSessionID, "testuser", "password")
	processLoginSrvResponse(conn, 5)

	fmt.Println("\n  ✅ LoginSrv 测试完成")
	waitEnter(reader)
}

func testM2Server(reader *bufio.Reader) {
	fmt.Println("\n═══════════════════════════════════════════════════════")
	fmt.Println("  测试3: M2Server (16000)")
	fmt.Println("═══════════════════════════════════════════════════════")

	fmt.Println("\n[3.1] 连接到 M2Server...")
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", M2ServerHost, M2ServerPort), 5*time.Second)
	if err != nil {
		fmt.Printf("  ⚠️  M2Server 未运行: %v\n", err)
		fmt.Println("  提示: 请先启动 M2Server 再进行测试")
		return
	}
	defer conn.Close()
	fmt.Println("  ✅ 已连接")

	fmt.Println("\n[3.2] 协议握手 (CM_PROTOCOL 2000)...")
	sendProtocolM2(conn)
	time.Sleep(200 * time.Millisecond)
	processM2ServerResponse(conn, 2)
	waitEnter(reader)

	fmt.Println("\n[3.3] 查询角色 (CM_QUERYCHR 100)...")
	sendQueryCharM2(conn)
	processM2ServerResponse(conn, 3)
	waitEnter(reader)

	fmt.Println("\n[3.4] 选择角色 (CM_SELCHR 103)...")
	sendSelectCharM2(conn, "Hero")
	processM2ServerResponse(conn, 5)
	waitEnter(reader)

	fmt.Println("\n[3.5] 移动测试 - 转身 (CM_TURN 3010)...")
	sendMovementM2(conn, 3010, 0, 0)
	processM2ServerResponse(conn, 2)
	waitEnter(reader)

	fmt.Println("\n[3.6] 移动测试 - 走路 (CM_WALK 3011)...")
	sendMovementM2(conn, 3011, 100, 100)
	processM2ServerResponse(conn, 2)
	waitEnter(reader)

	fmt.Println("\n[3.7] 移动测试 - 跑步 (CM_RUN 3013)...")
	sendMovementM2(conn, 3013, 101, 101)
	processM2ServerResponse(conn, 2)
	waitEnter(reader)

	fmt.Println("\n[3.8] 攻击测试 (CM_HIT 3014)...")
	sendMovementM2(conn, 3014, 0, 0)
	processM2ServerResponse(conn, 2)

	fmt.Println("\n  ✅ M2Server 测试完成")
}

// ========== LoginGate 消息 ==========

func sendNewConnection(conn net.Conn) {
	packet := "%N1/127.0.0.1$"
	n, err := conn.Write([]byte(packet))
	if err != nil {
		fmt.Printf("  ❌ 发送失败: %v\n", err)
		return
	}
	fmt.Printf("  发送: %s (%d bytes)\n", packet, n)
}

func sendMessage(conn net.Conn, ident int, body string) {
	msg := make([]byte, 12)
	binary.LittleEndian.PutUint32(msg[0:4], 0)
	binary.LittleEndian.PutUint16(msg[4:6], uint16(ident))

	data := msg
	if body != "" {
		data = make([]byte, 12+len(body))
		copy(data, msg)
		copy(data[12:], body)
	}

	encoded := encode6Bit(data)
	packet := fmt.Sprintf("#1%s!", encoded)
	n, err := conn.Write([]byte(packet))
	if err != nil {
		fmt.Printf("  ❌ 发送失败: %v\n", err)
		return
	}
	fmt.Printf("  发送: ident=%d (%d bytes)\n", ident, n)
}

func sendSelectServer(conn net.Conn, serverIndex int) {
	body := fmt.Sprintf("%d", serverIndex)
	msg := make([]byte, 12)
	binary.LittleEndian.PutUint32(msg[0:4], 0)
	binary.LittleEndian.PutUint16(msg[4:6], 104)
	binary.LittleEndian.PutUint16(msg[10:12], 4)

	data := make([]byte, 12+len(body))
	copy(data, msg)
	copy(data[12:], body)

	encoded := encode6Bit(data)
	packet := fmt.Sprintf("#1%s!", encoded)
	n, err := conn.Write([]byte(packet))
	if err != nil {
		fmt.Printf("  ❌ 发送失败: %v\n", err)
		return
	}
	fmt.Printf("  发送: CM_SELECTSERVER, index=%d (%d bytes)\n", serverIndex, n)
}

func sendLogin(conn net.Conn, account, password string) {
	body := account + "/" + password
	msg := make([]byte, 12)
	binary.LittleEndian.PutUint32(msg[0:4], 0)
	binary.LittleEndian.PutUint16(msg[4:6], 2001)
	binary.LittleEndian.PutUint16(msg[10:12], uint16(len(body)))

	data := make([]byte, 12+len(body))
	copy(data, msg)
	copy(data[12:], body)

	encoded := encode6Bit(data)
	packet := fmt.Sprintf("#1%s!", encoded)
	n, err := conn.Write([]byte(packet))
	if err != nil {
		fmt.Printf("  ❌ 发送失败: %v\n", err)
		return
	}
	fmt.Printf("  发送: CM_IDPASSWORD, account=%s (%d bytes)\n", account, n)
}

// ========== LoginSrv 消息 ==========

func sendNewConnectionLoginSrv(conn net.Conn) {
	packet := "%N1/127.0.0.1$"
	n, err := conn.Write([]byte(packet))
	if err != nil {
		fmt.Printf("  ❌ 发送失败: %v\n", err)
		return
	}
	fmt.Printf("  发送: %s (%d bytes)\n", packet, n)
}

func sendQueryServerLoginSrv(conn net.Conn, sessionID int) {
	msg := make([]byte, 12)
	binary.LittleEndian.PutUint16(msg[4:6], 107)
	binary.LittleEndian.PutUint16(msg[10:12], 0)

	encoded := encode6Bit(msg)
	packet := fmt.Sprintf("#%d%s!", sessionID, encoded)
	fullPacket := fmt.Sprintf("%%%d/%s$", sessionID, packet)

	n, err := conn.Write([]byte(fullPacket))
	if err != nil {
		fmt.Printf("  ❌ 发送失败: %v\n", err)
		return
	}
	fmt.Printf("  发送: CM_QUERYSERVERNAME (%d bytes)\n", n)
}

func sendSelectServerLoginSrv(conn net.Conn, sessionID int, serverIndex int) {
	body := fmt.Sprintf("%d", serverIndex)
	msg := make([]byte, 12)
	binary.LittleEndian.PutUint16(msg[4:6], 104)
	binary.LittleEndian.PutUint16(msg[10:12], 4)

	encoded := encode6Bit(msg)
	packet := fmt.Sprintf("#%d%s%s!", sessionID, encoded, body)
	fullPacket := fmt.Sprintf("%%%d/%s$", sessionID, packet)

	n, err := conn.Write([]byte(fullPacket))
	if err != nil {
		fmt.Printf("  ❌ 发送失败: %v\n", err)
		return
	}
	fmt.Printf("  发送: CM_SELECTSERVER (%d bytes)\n", n)
}

func sendLoginLoginSrv(conn net.Conn, sessionID int, account, password string) {
	body := account + "/" + password
	data := encodeString(body)

	msg := make([]byte, 12)
	binary.LittleEndian.PutUint16(msg[4:6], 2001)
	binary.LittleEndian.PutUint16(msg[10:12], uint16(len(body)))

	encoded := encode6Bit(msg)
	packet := fmt.Sprintf("#%d%s%s!", sessionID, encoded, data)
	fullPacket := fmt.Sprintf("%%%d/%s$", sessionID, packet)

	n, err := conn.Write([]byte(fullPacket))
	if err != nil {
		fmt.Printf("  ❌ 发送失败: %v\n", err)
		return
	}
	fmt.Printf("  发送: CM_IDPASSWORD (%d bytes)\n", n)
}

// ========== M2Server 消息 ==========

func sendProtocolM2(conn net.Conn) {
	msg := make([]byte, 14)
	binary.LittleEndian.PutUint16(msg[4:6], 2000)
	sendM2Packet(conn, msg)
	fmt.Println("  发送: CM_PROTOCOL (2000)")
}

func sendQueryCharM2(conn net.Conn) {
	msg := make([]byte, 14)
	binary.LittleEndian.PutUint16(msg[4:6], 100)
	sendM2Packet(conn, msg)
	fmt.Println("  发送: CM_QUERYCHR (100)")
}

func sendSelectCharM2(conn net.Conn, name string) {
	nameBytes := make([]byte, 30)
	copy(nameBytes, []byte(name))

	msg := make([]byte, 14+len(nameBytes))
	binary.LittleEndian.PutUint16(msg[4:6], 103)
	copy(msg[14:], nameBytes)
	sendM2Packet(conn, msg)
	fmt.Printf("  发送: CM_SELCHR (103), name=%s\n", name)
}

func sendMovementM2(conn net.Conn, msgID uint16, x, y uint16) {
	msg := make([]byte, 14)
	binary.LittleEndian.PutUint16(msg[4:6], msgID)
	binary.LittleEndian.PutUint16(msg[6:8], x)
	binary.LittleEndian.PutUint16(msg[8:10], y)
	sendM2Packet(conn, msg)

	name := "UNKNOWN"
	switch msgID {
	case 3010:
		name = "CM_TURN"
	case 3011:
		name = "CM_WALK"
	case 3013:
		name = "CM_RUN"
	case 3014:
		name = "CM_HIT"
	}
	fmt.Printf("  发送: %s (%d), x=%d, y=%d\n", name, msgID, x, y)
}

func sendM2Packet(conn net.Conn, data []byte) {
	packet := make([]byte, 6+len(data))
	binary.LittleEndian.PutUint32(packet[0:4], 0xAA55AA55)
	binary.LittleEndian.PutUint16(packet[4:6], uint16(len(data)))
	copy(packet[6:], data)
	conn.Write(packet)
}

// ========== 响应处理 ==========

func processResponses(conn net.Conn, timeoutSec int) {
	conn.SetReadDeadline(time.Now().Add(time.Duration(timeoutSec) * time.Second))
	buf := make([]byte, 8192)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Printf("  ⏱ 等待超时 (%ds)\n", timeoutSec)
			} else {
				fmt.Printf("  读取结束: %v\n", err)
			}
			break
		}

		data := string(buf[:n])
		fmt.Printf("  收到 (%d bytes)\n", n)

		for len(data) > 0 {
			idx := strings.Index(data, "!")
			if idx < 0 {
				break
			}

			packet := data[:idx+1]
			data = data[idx+1:]

			if len(packet) < 2 || packet[0] != '#' {
				continue
			}

			decoded := decode6Bit(packet[2 : len(packet)-1])
			if len(decoded) >= 12 {
				parseAndPrintMessage(decoded)
			}
		}
	}
}

var loginSrvSessionID int

func processLoginSrvResponse(conn net.Conn, timeoutSec int) {
	conn.SetReadDeadline(time.Now().Add(time.Duration(timeoutSec) * time.Second))
	buf := make([]byte, 4096)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Printf("  ⏱ 等待超时 (%ds)\n", timeoutSec)
			} else {
				fmt.Printf("  读取结束: %v\n", err)
			}
			break
		}

		data := string(buf[:n])
		fmt.Printf("  收到 (%d bytes): %x\n", n, buf[:n])

		if strings.HasPrefix(data, "%") {
			parts := strings.SplitN(data, "/", 2)
			if len(parts) >= 2 {
				packet := strings.TrimSuffix(parts[1], "$")
				if strings.HasPrefix(packet, "#") && len(packet) > 2 {
					slashIdx := strings.Index(packet[1:], "/")
					if slashIdx > 0 {
						sessionIDStr := packet[1:slashIdx+1]
						fmt.Sscanf(sessionIDStr, "%d", &loginSrvSessionID)
						fmt.Printf("  Session ID: %d\n", loginSrvSessionID)
						encoded := packet[slashIdx+2 : len(packet)-1]
						decoded := decode6Bit(encoded)
						parseAndPrintMessage(decoded)
					}
				}
			}
		}
	}
}

func processM2ServerResponse(conn net.Conn, timeoutSec int) {
	conn.SetReadDeadline(time.Now().Add(time.Duration(timeoutSec) * time.Second))
	buf := make([]byte, 8192)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Printf("  ⏱ 等待超时 (%ds)\n", timeoutSec)
			} else {
				fmt.Printf("  读取结束: %v\n", err)
			}
			break
		}

		pos := 0
		for pos < n {
			if pos+6 > n {
				break
			}

			header := binary.LittleEndian.Uint32(buf[pos : pos+4])
			if header != 0xAA55AA55 {
				pos++
				continue
			}

			length := binary.LittleEndian.Uint16(buf[pos+4 : pos+6])
			if pos+6+int(length) > n {
				break
			}

			payload := buf[pos+6 : pos+6+int(length)]
			parseAndPrintMessage(payload)

			pos += 6 + int(length)
		}
	}
}

func parseAndPrintMessage(data []byte) {
	if len(data) < 6 {
		return
	}

	recog := binary.LittleEndian.Uint32(data[0:4])
	ident := binary.LittleEndian.Uint16(data[4:6])

	var param, tag, series uint16
	if len(data) >= 8 {
		param = binary.LittleEndian.Uint16(data[6:8])
	}
	if len(data) >= 10 {
		tag = binary.LittleEndian.Uint16(data[8:10])
	}
	if len(data) >= 12 {
		series = binary.LittleEndian.Uint16(data[10:12])
	}

	var body string
	if len(data) > 12 {
		body = string(data[12:])
	}

	fmt.Printf("  <- Ident=%d, Recog=%d, Param=%d, Tag=%d, Series=%d\n",
		ident, recog, param, tag, series)

	if body != "" {
		fmt.Printf("    Body: %s\n", body)
	}

	switch ident {
	case 537:
		fmt.Printf("    => SM_SERVERNAME (服务器列表)\n")
	case 529:
		fmt.Printf("    => SM_PASSOK_SELECTSERVER (密码验证成功)\n")
	case 119:
		if len(data) > 14 {
			result := data[14]
			switch result {
			case 0:
				fmt.Printf("    => SM_LOGIN_SUCCESS (登录成功)\n")
			case 1:
				fmt.Printf("    => SM_LOGIN_FAIL (数据库错误)\n")
			case 2:
				fmt.Printf("    => SM_LOGIN_FAIL (账号不存在)\n")
			case 3:
				fmt.Printf("    => SM_LOGIN_FAIL (密码错误)\n")
			case 4:
				fmt.Printf("    => SM_LOGIN_FAIL (账号已在线)\n")
			default:
				fmt.Printf("    => SM_LOGIN_FAIL (code=%d)\n", result)
			}
		} else {
			fmt.Printf("    => SM_LOGIN_SUCCESS (登录成功)\n")
		}
	case 530:
		fmt.Printf("    => SM_SELECTSERVER_OK (服务器选择成功)\n")
	case 520:
		fmt.Printf("    => SM_QUERYCHR (角色列表)\n")
	case 525:
		fmt.Printf("    => SM_STARTPLAY (开始游戏)\n")
	case 50:
		fmt.Printf("    => SM_LOGON (登录成功)\n")
	case 51:
		fmt.Printf("    => SM_NEWMAP (新地图)\n")
	case 52:
		fmt.Printf("    => SM_ABILITY (属性)\n")
	case 20002:
		fmt.Printf("    => SM_SERVERVERSION (服务器版本)\n")
	case 10, 11, 13, 14:
		name := "UNKNOWN"
		switch ident {
		case 10:
			name = "SM_TURN"
		case 11:
			name = "SM_WALK"
		case 13:
			name = "SM_RUN"
		case 14:
			name = "SM_HIT"
		}
		fmt.Printf("    => %s (移动/攻击广播)\n", name)
	}
}

func waitEnter(reader *bufio.Reader) {
	// Skip waiting, auto continue for testing
	fmt.Println()
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

func encodeString(s string) string {
	return encode6Bit([]byte(s))
}

func clearScreen() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}