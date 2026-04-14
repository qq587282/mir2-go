package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("=== 完整登录测试客户端 ===")

	// 阶段1: 登录服务器
	fmt.Println("\n=== 阶段1: 登录服务器 ===")
	loginConn := connectServer("127.0.0.1", 15500)
	if loginConn == nil {
		return
	}

	sendLoginFormat := func(conn net.Conn, msg string) {
		packet := fmt.Sprintf("%%%s$", msg)
		conn.Write([]byte(packet))
	}

	sendLoginFormat(loginConn, "N1/127.0.0.1")
	fmt.Println("发送: N1/127.0.0.1")
	time.Sleep(200 * time.Millisecond)
	recvAll(loginConn)

	time.Sleep(200 * time.Millisecond)
	sendLoginFormat(loginConn, "D0/#0<<<<<B1<<<<<<<1!")  // SelectServer
	fmt.Println("发送: 选择服务器")
	time.Sleep(200 * time.Millisecond)
	recvAll(loginConn)

	time.Sleep(200 * time.Millisecond)
	sendLoginFormat(loginConn, "D0/#0<<<<<I@C<<<<<=X<YBQoYCQoUSDmH_HkXBAoXsYkXbLmH_H!")  // Login
	fmt.Println("发送: 登录")
	time.Sleep(200 * time.Millisecond)
	recvAll(loginConn)

	loginConn.Close()
	fmt.Println("登录服务器关闭")

	// 阶段2: 通过RunGate连接M2Server
	fmt.Println("\n=== 阶段2: 通过RunGate连接M2Server ===")
	
	rgConn := connectServer("127.0.0.1", 7000)
	if rgConn == nil {
		return
	}
	defer rgConn.Close()

	sendRungateFormat := func(conn net.Conn, msg string) {
		conn.Write([]byte(msg))
	}

	recvRungate := func(conn net.Conn) {
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("接收结束: %v\n", err)
			return
		}
		data := buf[:n]
		fmt.Printf("收到 %d 字节: %x\n", n, data)
		if n > 20 {
			msg := decode6Bit(string(data[1:len(data)-1]))
			if len(msg) >= 14 {
				ident := binary.LittleEndian.Uint16(msg[4:6])
				recog := int32(binary.LittleEndian.Uint32(msg[0:4]))
				fmt.Printf("  -> Ident=%d, Recog=%d\n", ident, recog)
			}
		}
	}

	time.Sleep(500 * time.Millisecond)

	// 发送角色查询
	sendRungateFormat(rgConn, "#1<<<<<=@><<<<<<<!")
	fmt.Println("发送: CM_QUERYCHR")
	recvRungate(rgConn)

	time.Sleep(500 * time.Millisecond)

	// 发送角色选择 (选择不存在的角色，服务器会创建临时角色)
	nameBytes := make([]byte, 14)
	copy(nameBytes, []byte("Hero"))
	encodedName := encode6Bit(nameBytes)
	sendRungateFormat(rgConn, fmt.Sprintf("#1<<<<<I@C<<<<<=X<%s!", encodedName))
	fmt.Println("发送: CM_SELCHR")
	time.Sleep(2 * time.Second)
	recvRungate(rgConn)
	recvRungate(rgConn)
	recvRungate(rgConn)

	time.Sleep(500 * time.Millisecond)

	// 发送转身
	sendRungateFormat(rgConn, "#1<<<<<=D><<<<<<<!")
	fmt.Println("发送: CM_TURN")
	recvRungate(rgConn)

	time.Sleep(200 * time.Millisecond)

	// 发送走路
	sendRungateFormat(rgConn, "#1<<<<<=E><<<<<<<!")
	fmt.Println("发送: CM_WALK")
	recvRungate(rgConn)

	fmt.Println("\n=== 测试完成 ===")
}

func connectServer(ip string, port int) net.Conn {
	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		fmt.Printf("连接失败 %s: %v\n", addr, err)
		return nil
	}
	fmt.Printf("已连接到 %s\n", addr)
	return conn
}

func recvAll(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	buf := make([]byte, 4096)
	for {
		n, err := conn.Read(buf)
		if err != nil {
			break
		}
		if n > 0 {
			fmt.Printf("收到 %d 字节\n", n)
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
