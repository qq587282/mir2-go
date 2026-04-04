package main

import (
	"fmt"
	"net"
	"os"
	"time"
)

func main() {
	fmt.Println("=== Raw TCP Server Test ===")

	ln, err := net.Listen("tcp", "0.0.0.0:15500")
	if err != nil {
		fmt.Printf("Listen error: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Listening on 0.0.0.0:15500")

	go func() {
		for {
			fmt.Println("Waiting for Accept()...")
			conn, err := ln.Accept()
			if err != nil {
				fmt.Printf("Accept error: %v\n", err)
				continue
			}
			fmt.Printf("Accepted: %s\n", conn.RemoteAddr())

			go handleConn(conn)
		}
	}()

	time.Sleep(30 * time.Second)
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	fmt.Printf("Handling connection from %s\n", conn.RemoteAddr())

	buf := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("Read error: %v\n", err)
		return
	}

	fmt.Printf("Received %d bytes: %x\n", n, buf[:n])

	conn.Write([]byte{0x01, 0x00, 0x00})
}
