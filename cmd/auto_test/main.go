package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("=== Mir2 自动流程测试 ===")

	start := time.Now()
	testLoginServer()
	testGameServer()

	fmt.Printf("\n=== 测试完成 (耗时: %v) ===\n", time.Since(start))
}

func testLoginServer() {
	fmt.Println("\n[测试1] LoginSrv (15500)...")

	conn, err := net.DialTimeout("tcp", "127.0.0.1:15500", 5*time.Second)
	if err != nil {
		fmt.Println("  连接失败:", err)
		return
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	conn.Write([]byte("%N1/127.0.0.1$"))

	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)
	if n > 0 {
		fmt.Printf("  握手: %d bytes OK\n", n)
	}

	buf = make([]byte, 4096)
	n, _ = conn.Read(buf)
	if n > 0 {
		ident := binary.LittleEndian.Uint16(buf[4:6])
		fmt.Printf("  服务器列表: ID=%d OK\n", ident)
	}
}

func testGameServer() {
	fmt.Println("\n[测试2] RunGate->M2Server (7200)...")

	conn, err := net.DialTimeout("tcp", "127.0.0.1:7200", 5*time.Second)
	if err != nil {
		fmt.Println("  连接失败:", err)
		return
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))

	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)
	if n > 0 {
		fmt.Printf("  配置: %d bytes OK\n", n)
	}

	msg := make([]byte, 12)
	binary.LittleEndian.PutUint32(msg[0:4], 0)
	binary.LittleEndian.PutUint16(msg[4:6], 100)
	encoded := encode6Bit(msg)
	conn.Write([]byte("#1" + encoded + "!"))

	buf = make([]byte, 4096)
	n, _ = conn.Read(buf)
	if n > 0 {
		ident := binary.LittleEndian.Uint16(buf[4:6])
		fmt.Printf("  角色查询: ID=%d OK\n", ident)
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