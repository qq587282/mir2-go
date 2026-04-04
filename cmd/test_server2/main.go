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

	fmt.Println("已连接")

	msg := "#1<<<<<Bh<<<<<<<<<!"
	conn.Write([]byte(msg))
	fmt.Println("发送:", msg)

	time.Sleep(500 * time.Millisecond)

	buf := make([]byte, 8192)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("读取错误:", err)
	} else {
		data := string(buf[:n])
		fmt.Printf("收到 (%d bytes): %s\n", n, data)
		
		for i := 0; i < n; i++ {
			if data[i] == '!' {
				packet := data[:i+1]
				fmt.Println("Packet:", packet)
				
				if len(packet) > 1 && packet[0] == '#' {
					code := int(packet[1] - '0')
					encoded := packet[2:len(packet)-1]
					raw := decode6Bit(encoded)
					fmt.Printf("解码 (code=%d, %d bytes): %x\n", code, len(raw), raw)
					if len(raw) >= 12 {
						ident := binary.LittleEndian.Uint16(raw[4:6])
						series := binary.LittleEndian.Uint16(raw[10:12])
						fmt.Printf("Ident=%d, Series=%d\n", ident, series)
						if len(raw) > 12 {
							fmt.Printf("Body: '%s'\n", string(raw[12:]))
						}
					}
				}
				data = data[i+1:]
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