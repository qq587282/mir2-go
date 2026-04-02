package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"time"
)

const (
	SERVER_HOST = "127.0.0.1"
	SERVER_PORT = 7200
	BUFFER_SIZE = 8192
)

type Client struct {
	conn      net.Conn
	reader    *bufio.Reader
	account   string
	sessionID int32
}

func main() {
	fmt.Println("Mir2 Client Test Tool")
	fmt.Println("=====================")
	
	client := &Client{}
	
	fmt.Println("Connecting to server...")
	err := client.Connect(SERVER_HOST, SERVER_PORT)
	if err != nil {
		fmt.Printf("Connection failed: %v\n", err)
		return
	}
	defer client.conn.Close()
	
	fmt.Println("Connected! Testing protocol...")
	
	client.testProtocol()
	
	fmt.Println("\nTesting login...")
	client.testLogin("testuser", "testpass")
	
	fmt.Println("\nTesting character query...")
	client.testQueryChar()
	
	fmt.Println("\nTesting character selection...")
	client.testSelectChar("TestChar")
	
	fmt.Println("\nTesting movement...")
	client.testMovement()
	
	fmt.Println("\nAll tests completed!")
	fmt.Println("Press Enter to exit...")
	bufio.NewReader(os.Stdin).ReadLine()
}

func (c *Client) Connect(host string, port int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	
	c.conn = conn
	c.reader = bufio.NewReaderSize(conn, BUFFER_SIZE)
	
	go c.receiveLoop()
	
	return nil
}

func (c *Client) receiveLoop() {
	buf := make([]byte, BUFFER_SIZE)
	for {
		n, err := c.conn.Read(buf)
		if err != nil {
			fmt.Println("Disconnected:", err)
			return
		}
		
		c.processPacket(buf[:n])
	}
}

func (c *Client) processPacket(data []byte) {
	if len(data) < 6 {
		return
	}
	
	header := binary.LittleEndian.Uint32(data[0:4])
	if header != 0xAA55AA55 {
		fmt.Printf("Invalid header: %x\n", header)
		return
	}
	
	length := binary.LittleEndian.Uint16(data[2:4])
	if len(data) < int(4+length) {
		return
	}
	
	packet := data[4 : 4+length]
	
	if len(packet) >= 14 {
		ident := binary.LittleEndian.Uint16(packet[4:6])
		recog := int32(binary.LittleEndian.Uint32(packet[0:4]))
		param := binary.LittleEndian.Uint16(packet[6:8])
		tag := binary.LittleEndian.Uint16(packet[8:10])
		series := binary.LittleEndian.Uint16(packet[10:12])
		
		fmt.Printf("Recv: Ident=%d, Recog=%d, Param=%d, Tag=%d, Series=%d\n", 
			ident, recog, param, tag, series)
		
		body := packet[14:]
		if len(body) > 0 {
			fmt.Printf("  Body (%d bytes): %v\n", len(body), body[:min(20, len(body))])
		}
	} else {
		fmt.Printf("Recv packet: %v\n", packet)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (c *Client) testProtocol() {
	fmt.Println("Sending CM_PROTOCOL (2000)...")
	c.sendMessage(2000, 0, 0, 0, 0)
	time.Sleep(100 * time.Millisecond)
}

func (c *Client) testLogin(account, password string) {
	fmt.Printf("Sending CM_IDPASSWORD: account=%s, password=%s\n", account, password)
	
	body := make([]byte, 62)
	copy(body[0:30], []byte(account))
	copy(body[30:60], []byte(password))
	body[61] = 0
	
	c.sendMessageWithBody(2001, 0, 0, 0, 0, body)
	time.Sleep(100 * time.Millisecond)
}

func (c *Client) testQueryChar() {
	fmt.Println("Sending CM_QUERYCHR (100)...")
	c.sendMessage(100, 0, 0, 0, 0)
	time.Sleep(100 * time.Millisecond)
}

func (c *Client) testSelectChar(name string) {
	fmt.Printf("Sending CM_SELCHR (103): name=%s\n", name)
	
	body := make([]byte, 30)
	copy(body[0:len(name)], []byte(name))
	
	c.sendMessageWithBody(103, 0, 0, 0, 0, body)
	time.Sleep(100 * time.Millisecond)
}

func (c *Client) testMovement() {
	fmt.Println("Testing movement commands...")
	
	tests := []struct {
		name string
		ident uint16
		param uint16
		tag   uint16
	}{
		{"CM_TURN", 3010, 0, 0},
		{"CM_WALK", 3011, 0, 0},
		{"CM_RUN", 3013, 0, 0},
		{"CM_HIT", 3014, 0, 0},
	}
	
	for _, t := range tests {
		fmt.Printf("Sending %s (%d)...\n", t.name, t.ident)
		c.sendMessage(t.ident, int32(t.param), t.tag, 0, 0)
		time.Sleep(50 * time.Millisecond)
	}
}

func (c *Client) sendMessage(ident uint16, recog int32, param, tag, series uint16) {
	msg := make([]byte, 14)
	binary.LittleEndian.PutUint32(msg[0:4], uint32(recog))
	binary.LittleEndian.PutUint16(msg[4:6], ident)
	binary.LittleEndian.PutUint16(msg[6:8], param)
	binary.LittleEndian.PutUint16(msg[8:10], tag)
	binary.LittleEndian.PutUint16(msg[10:12], series)
	
	c.sendPacket(msg)
}

func (c *Client) sendMessageWithBody(ident uint16, recog int32, param, tag, series uint16, body []byte) {
	msg := make([]byte, 14+len(body))
	binary.LittleEndian.PutUint32(msg[0:4], uint32(recog))
	binary.LittleEndian.PutUint16(msg[4:6], ident)
	binary.LittleEndian.PutUint16(msg[6:8], param)
	binary.LittleEndian.PutUint16(msg[8:10], tag)
	binary.LittleEndian.PutUint16(msg[10:12], series)
	
	copy(msg[14:], body)
	
	c.sendPacket(msg)
}

func (c *Client) sendPacket(data []byte) {
	packet := make([]byte, 4+len(data))
	binary.LittleEndian.PutUint32(packet[0:4], 0xAA55AA55)
	copy(packet[4:], data)
	
	n, err := c.conn.Write(packet)
	if err != nil {
		fmt.Printf("Send failed: %v\n", err)
		return
	}
	
	fmt.Printf("Sent %d bytes\n", n)
}