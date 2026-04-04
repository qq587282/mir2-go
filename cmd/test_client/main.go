package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:7000")
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}
	defer conn.Close()

	fmt.Println("=== 传奇2客户端模拟器 ===")
	fmt.Println("已连接到LoginGate (127.0.0.1:7000)")

	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n[1] 发送 CM_QUERYSERVERNAME (ident=107)...")
	sendMessage(conn, 107, "")

	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n[2] 等待服务器响应...")
	processResponses(conn)

	time.Sleep(1 * time.Second)

	fmt.Println("\n[3] 发送 CM_SELECTSERVER 选择第一个服务器...")
	sendMessage(conn, 104, "Server1")

	time.Sleep(500 * time.Millisecond)

	fmt.Println("\n[4] 等待服务器响应...")
	processResponses(conn)

	time.Sleep(1 * time.Second)

	fmt.Println("\n[5] 发送 CM_IDPASSWORD (账号: test, 密码: test)...")
	sendLogin(conn, "test", "test")

	time.Sleep(1 * time.Second)

	fmt.Println("\n[6] 等待服务器响应...")
	processResponses(conn)

	fmt.Println("\n=== 测试完成 ===")
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
	conn.Write([]byte(packet))
	fmt.Printf("    发送: ident=%d, body='%s'\n", ident, body)
}

func sendLogin(conn net.Conn, account, password string) {
	body := account + "/" + password
	sendMessage(conn, 2001, body)
}

func processResponses(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 8192)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}

		data := string(buf[:n])
		fmt.Printf("    收到原始数据 (%d bytes): %x\n", n, buf[:n])

		for len(data) > 0 {
			idx := -1
			for i := 0; i < len(data); i++ {
				if data[i] == '!' {
					idx = i
					break
				}
			}
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
				ident := binary.LittleEndian.Uint16(decoded[4:6])
				series := binary.LittleEndian.Uint16(decoded[10:12])

				var body string
				if len(decoded) > 12 {
					body = string(decoded[12:])
				}

				fmt.Printf("    -> ident=%d (0x%x), series=%d, body='%s'\n", ident, ident, series, body)

				switch ident {
				case 537:
					fmt.Println("    [SM_SERVERNAME] 服务器列表")
				case 529:
					fmt.Println("    [SM_PASSOK_SELECTSERVER] 密码验证成功")
				case 119:
					fmt.Println("    [SM_LOGIN_SUCCESS] 登录成功")
				case 530:
					fmt.Println("    [SM_SELECTSERVER_OK] 服务器选择成功")
				default:
					fmt.Printf("    [未知消息 %d]\n", ident)
				}
			}
		}
	}
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
