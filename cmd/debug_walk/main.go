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
	fmt.Println("=== Debug Walk Test ===")

	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", M2ServerHost, M2ServerPort), 5*time.Second)
	if err != nil {
		fmt.Printf("连接失败: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Println("已连接到 M2Server")

	time.Sleep(500 * time.Millisecond)

	// Send protocol
	sendPacket(conn, 2000, nil)
	time.Sleep(200 * time.Millisecond)
	processResponse(conn)

	// Select character
	sendSelectChar(conn, "TestChar")
	time.Sleep(500 * time.Millisecond)
	processResponse(conn)

	// Debug: Send walk and print raw response
	fmt.Println("\n[DEBUG] 发送 CM_WALK...")
	sendWalk(conn, 289, 618)

	time.Sleep(1 * time.Second)

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 8192)
	n, err := conn.Read(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Println("超时")
		} else {
			fmt.Printf("读取: %v\n", err)
		}
	}

	fmt.Printf("\n原始数据 (%d bytes):\n", n)
	for i := 0; i < n; i += 16 {
		hex := ""
		for j := i; j < i+16 && j < n; j++ {
			hex += fmt.Sprintf("%02x ", buf[j])
		}
		fmt.Printf("  %04x: %s\n", i, hex)
	}

	// Parse and print
	pos := 0
	for pos < n {
		if pos+6 > n {
			break
		}
		header := binary.LittleEndian.Uint32(buf[pos : pos+4])
		fmt.Printf("\n包头: 0x%08x\n", header)
		if header != 0xAA55AA55 {
			pos++
			continue
		}
		length := binary.LittleEndian.Uint16(buf[pos+4 : pos+6])
		fmt.Printf("长度: %d\n", length)
		if pos+6+int(length) > n {
			break
		}
		payload := buf[pos+6 : pos+6+int(length)]
		fmt.Printf("载荷 (%d bytes): ", len(payload))
		for _, b := range payload {
			fmt.Printf("%02x ", b)
		}
		fmt.Println()

		if len(payload) >= 6 {
			recog := binary.LittleEndian.Uint32(payload[0:4])
			ident := binary.LittleEndian.Uint16(payload[4:6])
			param := binary.LittleEndian.Uint16(payload[6:8])
			tag := binary.LittleEndian.Uint16(payload[8:10])
			fmt.Printf("  Recog=%d, Ident=%d, Param=%d, Tag=%d\n", recog, ident, param, tag)
		}

		pos += 6 + int(length)
	}
}

func sendPacket(conn net.Conn, ident uint16, body []byte) {
	msg := make([]byte, 14)
	binary.LittleEndian.PutUint16(msg[4:6], ident)
	if body != nil {
		newMsg := make([]byte, 14+len(body))
		copy(newMsg, msg)
		copy(newMsg[14:], body)
		msg = newMsg
	}
	sendM2Packet(conn, msg)
}

func sendSelectChar(conn net.Conn, name string) {
	nameBytes := make([]byte, 30)
	copy(nameBytes, []byte(name))
	sendPacket(conn, 103, nameBytes)
}

func sendWalk(conn net.Conn, x, y int) {
	msg := make([]byte, 14)
	binary.LittleEndian.PutUint16(msg[4:6], 3011)
	binary.LittleEndian.PutUint16(msg[6:8], uint16(x))
	binary.LittleEndian.PutUint16(msg[8:10], uint16(y))
	fmt.Printf("发送 CM_WALK: ident=%d, x=%d, y=%d\n", 3011, x, y)
	fmt.Printf("原始: ")
	for _, b := range msg {
		fmt.Printf("%02x ", b)
	}
	fmt.Println()
	sendM2Packet(conn, msg)
}

func sendM2Packet(conn net.Conn, data []byte) {
	packet := make([]byte, 6+len(data))
	binary.LittleEndian.PutUint32(packet[0:4], 0xAA55AA55)
	binary.LittleEndian.PutUint16(packet[4:6], uint16(len(data)))
	copy(packet[6:], data)
	conn.Write(packet)
}

func processResponse(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 8192)
	n, err := conn.Read(buf)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			fmt.Println("超时")
		}
		return
	}
	fmt.Printf("收到 %d bytes\n", n)
}
