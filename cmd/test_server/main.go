package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:7000")
	if err != nil {
		fmt.Println("连接失败:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("已连接")

	// 1. 发送 CM_QUERYSERVERNAME (107)
	msg := make([]byte, 12)
	binary.LittleEndian.PutUint16(msg[4:6], 107) // ident
	encoded := encode6Bit(msg)
	packet := fmt.Sprintf("#1%s!", encoded)
	fmt.Println("发送:", packet)
	conn.Write([]byte(packet))

	// 接收响应
	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)
	data := string(buf[:n])
	fmt.Println("收到:", data)
	fmt.Println("原始:", []byte(data))

	// 解码
	if len(data) > 0 && data[0] == '#' {
		parts := data[1:]
		if len(parts) >= 2 {
			code := int(parts[0] - '0')
			encoded := parts[1:]
			raw := decode6Bit(encoded)
			fmt.Printf("解码后 (code=%d): %x\n", code, raw)
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