package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

const (
	M2ServerHost = "127.0.0.1"
	M2ServerPort = 16000
)

func main() {
	fmt.Println("=== M2Server 移动指令测试 ===")
	fmt.Println()

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", M2ServerHost, M2ServerPort), 5*time.Second)
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Println("✅ 已连接到 M2Server")

	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n[步骤1] 发送协议握手 (CM_PROTOCOL 2000)...")
	sendProtocol(conn)
	time.Sleep(200 * time.Millisecond)
	processResponse(conn)

	fmt.Println("\n[步骤2] 查询角色 (CM_QUERYCHR 100)...")
	sendQueryChar(conn)
	time.Sleep(200 * time.Millisecond)
	processResponse(conn)

	fmt.Println("\n[步骤3] 选择角色 (CM_SELCHR 103)...")
	sendSelectChar(conn, "TestChar")
	time.Sleep(500 * time.Millisecond)
	processResponse(conn)

	fmt.Println("\n[步骤4] 转身 (CM_TURN 3010)...")
	sendTurn(conn, 2)
	time.Sleep(500 * time.Millisecond)
	processResponse(conn)

	fmt.Println("\n[步骤5] 走路 (CM_WALK 3011)...")
	sendWalk(conn, 289, 618)
	time.Sleep(500 * time.Millisecond)
	processResponse(conn)

	fmt.Println("\n[步骤6] 跑步 (CM_RUN 3013)...")
	sendRun(conn, 289, 619)
	time.Sleep(500 * time.Millisecond)
	processResponse(conn)

	fmt.Println("\n[步骤7] 攻击 (CM_HIT 3014)...")
	sendHit(conn)
	time.Sleep(500 * time.Millisecond)
	processResponse(conn)

	fmt.Println("\n=== 测试完成 ===")
}

func sendProtocol(conn net.Conn) {
	msg := make([]byte, 14)
	binary.LittleEndian.PutUint16(msg[4:6], 2000)
	sendM2Packet(conn, msg)
	fmt.Println("  发送: CM_PROTOCOL (2000)")
}

func sendQueryChar(conn net.Conn) {
	msg := make([]byte, 14)
	binary.LittleEndian.PutUint16(msg[4:6], 100)
	sendM2Packet(conn, msg)
	fmt.Println("  发送: CM_QUERYCHR (100)")
}

func sendSelectChar(conn net.Conn, name string) {
	nameBytes := make([]byte, 30)
	copy(nameBytes, []byte(name))

	msg := make([]byte, 14+len(nameBytes))
	binary.LittleEndian.PutUint16(msg[4:6], 103)
	copy(msg[14:], nameBytes)
	sendM2Packet(conn, msg)
	fmt.Printf("  发送: CM_SELCHR (103), name=%s\n", name)
}

func sendTurn(conn net.Conn, direction byte) {
	msg := make([]byte, 14)
	binary.LittleEndian.PutUint16(msg[4:6], 3010)
	binary.LittleEndian.PutUint16(msg[6:8], uint16(direction))
	sendM2Packet(conn, msg)
	fmt.Printf("  发送: CM_TURN (3010), direction=%d\n", direction)
}

func sendWalk(conn net.Conn, x, y int) {
	msg := make([]byte, 14)
	binary.LittleEndian.PutUint16(msg[4:6], 3011)
	binary.LittleEndian.PutUint16(msg[6:8], uint16(x))
	binary.LittleEndian.PutUint16(msg[8:10], uint16(y))
	sendM2Packet(conn, msg)
	fmt.Printf("  发送: CM_WALK (3011), x=%d, y=%d\n", x, y)
}

func sendRun(conn net.Conn, x, y int) {
	msg := make([]byte, 14)
	binary.LittleEndian.PutUint16(msg[4:6], 3013)
	binary.LittleEndian.PutUint16(msg[6:8], uint16(x))
	binary.LittleEndian.PutUint16(msg[8:10], uint16(y))
	sendM2Packet(conn, msg)
	fmt.Printf("  发送: CM_RUN (3013), x=%d, y=%d\n", x, y)
}

func sendHit(conn net.Conn) {
	msg := make([]byte, 14)
	binary.LittleEndian.PutUint16(msg[4:6], 3014)
	sendM2Packet(conn, msg)
	fmt.Println("  发送: CM_HIT (3014)")
}

func sendM2Packet(conn net.Conn, data []byte) {
	packet := make([]byte, 6+len(data))
	binary.LittleEndian.PutUint32(packet[0:4], 0xAA55AA55)
	binary.LittleEndian.PutUint16(packet[4:6], uint16(len(data)))
	copy(packet[6:], data)
	conn.Write(packet)
}

func processResponse(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 8192)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Println("  ⏱ 等待超时 (3s)")
			} else {
				fmt.Printf("  读取结束: %v\n", err)
			}
			return
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
	case 54:
		fmt.Printf("    => SM_MAPDESCRIPTION (地图描述)\n")
	case 20002:
		fmt.Printf("    => SM_SERVERCONFIG (服务器配置)\n")
	case 10:
		fmt.Printf("    => SM_TURN (转身广播)\n")
	case 11:
		fmt.Printf("    => SM_WALK (走路广播)\n")
	case 13:
		fmt.Printf("    => SM_RUN (跑步广播)\n")
	case 14:
		fmt.Printf("    => SM_HIT (攻击广播)\n")
	case 28:
		fmt.Printf("    => SM_MOVEFAIL (移动失败)\n")
	default:
		fmt.Printf("    => 未知消息 %d\n", ident)
	}
}
