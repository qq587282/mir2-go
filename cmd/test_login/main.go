package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:15500")
	if err != nil {
		fmt.Println("连接失败:", err)
		os.Exit(1)
	}
	defer conn.Close()

	fmt.Println("已连接LoginSrv")

	msg := "#1<<<<<Bh<<<<<<<<<!"

	packet := make([]byte, 6+len(msg))
	binary.LittleEndian.PutUint32(packet[0:4], 0xAA55AA55)
	binary.LittleEndian.PutUint16(packet[4:6], uint16(len(msg)))
	copy(packet[6:], msg)

	fmt.Printf("发送: %x\n", packet)
	conn.Write(packet)

	time.Sleep(500 * time.Millisecond)

	buf := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("读取错误:", err)
	} else {
		fmt.Printf("收到 (%d bytes): %x\n", n, buf[:n])
	}
}