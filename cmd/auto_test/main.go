package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

func main() {
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║         Mir2 Auto Test v1.0                        ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")

	loginConn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", "127.0.0.1", 7000))
	if err != nil {
		fmt.Printf("❌ 连接失败: %v\n", err)
		return
	}
	defer loginConn.Close()

	fmt.Println("✅ 已连接到 LoginGate")

	sendNewConnection(loginConn)
	time.Sleep(500 * time.Millisecond)

	processLoginGateResponse(loginConn, 3)

	fmt.Println("\n查询服务器列表...")
	sendMessage(loginConn, 107, "")
	time.Sleep(500 * time.Millisecond)
	processLoginGateResponse(loginConn, 3)

	fmt.Println("\n测试完成")
}

func sendNewConnection(conn net.Conn) {
	packet := "%N1/127.0.0.1$"
	conn.Write([]byte(packet))
	fmt.Printf("发送: %s\n", packet)
}

func sendMessage(conn net.Conn, ident int, body string) {
	msg := make([]byte, 12)
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
}

func processLoginGateResponse(conn net.Conn, timeoutSec int) {
	conn.SetReadDeadline(time.Now().Add(time.Duration(timeoutSec) * time.Second))
	buf := make([]byte, 8192)

	for {
		n, err := conn.Read(buf)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				fmt.Printf("⏱ 等待超时 (%ds)\n", timeoutSec)
			}
			break
		}

		data := string(buf[:n])
		fmt.Printf("收到 (%d bytes): %x\n", n, buf[:n])

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
				ident := binary.LittleEndian.Uint16(decoded[4:6])
				fmt.Printf("-> Ident=%d\n", ident)
			}
		}
	}
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