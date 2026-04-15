package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

func main() {
	fmt.Println("=== Mir2 游戏流程测试 ===")

	conn, err := net.DialTimeout("tcp", "127.0.0.1:7200", 5*time.Second)
	if err != nil {
		fmt.Println("连接失败:", err)
		return
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(20 * time.Second))
	fmt.Println("已连接到 RunGate (7200)")

	buf := make([]byte, 4096)
	n, _ := conn.Read(buf)
	fmt.Printf("[0] SM_SERVERCONFIG: %d bytes OK\n", n)

	time.Sleep(500 * time.Millisecond)
	testCharacterQuery(conn)
	testCharacterSelection(conn)
	testTurn(conn)
	testWalk(conn)
	testRun(conn)
	testHit(conn)
	testSpell(conn)
	testPickUp(conn)

	fmt.Println("\n=== 测试完成 ===")
}

func testCharacterQuery(conn net.Conn) {
	fmt.Println("\n[1] 角色查询 (CM_QUERYCHR 100)...")
	sendGameMsg(conn, 100, 0, 0)
	time.Sleep(500 * time.Millisecond)
	recvResponse(conn)
}

func testCharacterSelection(conn net.Conn) {
	fmt.Println("\n[2] 角色选择 (CM_SELCHR 103)...")
	name := "TestHero"
	nameBytes := []byte(name)
	for len(nameBytes) < 30 {
		nameBytes = append(nameBytes, 0)
	}

	msg := make([]byte, 14+30)
	binary.LittleEndian.PutUint32(msg[0:4], 0)
	binary.LittleEndian.PutUint16(msg[4:6], 103)
	binary.LittleEndian.PutUint16(msg[10:12], 30)
	copy(msg[14:], nameBytes)
	encoded := encode6Bit(msg)
	conn.Write([]byte("#1" + encoded + "!"))

	for i := 0; i < 5; i++ {
		conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		data := make([]byte, 4096)
		n, err := conn.Read(data)
		if err != nil || n == 0 {
			break
		}

		decoded := decodePacket(data[:n])
		if decoded == nil || len(decoded) < 6 {
			continue
		}

		ident := binary.LittleEndian.Uint16(decoded[4:6])
		recog := binary.LittleEndian.Uint32(decoded[0:4])
		fmt.Printf("    [%d] ID=%d, Recog=%d\n", i, ident, recog)

		switch ident {
		case 525:
			fmt.Println("    -> SM_STARTPLAY")
		case 50:
			fmt.Println("    -> SM_LOGON")
		case 51:
			fmt.Println("    -> SM_NEWMAP")
		}
	}
}

func testTurn(conn net.Conn) {
	fmt.Println("\n[3] 转身 (CM_TURN 3010)...")
	sendGameMsg(conn, 3010, 0, 0)
	recvResponse(conn)
}

func testWalk(conn net.Conn) {
	fmt.Println("\n[4] 走路 (CM_WALK 3011)...")
	sendGameMsg(conn, 3011, 300, 300)
	recvResponse(conn)
}

func testRun(conn net.Conn) {
	fmt.Println("\n[5] 跑步 (CM_RUN 3013)...")
	sendGameMsg(conn, 3013, 301, 301)
	recvResponse(conn)
}

func testHit(conn net.Conn) {
	fmt.Println("\n[6] 攻击 (CM_HIT 3014)...")
	sendGameMsg(conn, 3014, 0, 0)
	recvResponse(conn)
}

func testSpell(conn net.Conn) {
	fmt.Println("\n[7] 技能 (CM_SPELL 3015)...")
	sendGameMsg(conn, 3015, 0, 0)
	recvResponse(conn)
}

func testPickUp(conn net.Conn) {
	fmt.Println("\n[8] 拾取 (CM_PICKUP 3018)...")
	sendGameMsg(conn, 3018, 0, 0)
	recvResponse(conn)
}

func sendGameMsg(conn net.Conn, ident, x, y int) {
	msg := make([]byte, 12)
	binary.LittleEndian.PutUint32(msg[0:4], 0)
	binary.LittleEndian.PutUint16(msg[4:6], uint16(ident))
	binary.LittleEndian.PutUint16(msg[6:8], uint16(x)|uint16((y&0xFF)<<8))
	binary.LittleEndian.PutUint16(msg[8:10], 0)
	binary.LittleEndian.PutUint16(msg[10:12], 0)
	encoded := encode6Bit(msg)
	conn.Write([]byte("#1" + encoded + "!"))
}

func recvResponse(conn net.Conn) {
	conn.SetReadDeadline(time.Now().Add(3 * time.Second))
	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil || n == 0 {
		fmt.Println("    无响应")
		return
	}

	decoded := decodePacket(buf[:n])
	if decoded == nil || len(decoded) < 6 {
		fmt.Printf("    原始: %d bytes\n", n)
		return
	}

	ident := binary.LittleEndian.Uint16(decoded[4:6])
	recog := binary.LittleEndian.Uint32(decoded[0:4])
	fmt.Printf("    ID=%d, Recog=%d\n", ident, recog)

	switch ident {
	case 10:
		fmt.Println("    -> SM_TURN")
	case 11:
		fmt.Println("    -> SM_WALK")
	case 13:
		fmt.Println("    -> SM_RUN")
	case 14:
		fmt.Println("    -> SM_HIT")
	}
}

func decodePacket(data []byte) []byte {
	for i := 0; i < len(data); i++ {
		if data[i] == '#' {
			for j := i + 1; j < len(data); j++ {
				if data[j] == '!' && j-i > 2 {
					return decode6Bit(string(data[i+1 : j]))
				}
			}
		}
		if i+4 < len(data) && data[i] == 0x55 && data[i+1] == 0xAA {
			length := int(data[i+2])
			if i+4+length <= len(data) {
				return data[i+4 : i+4+length]
			}
		}
	}
	return nil
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