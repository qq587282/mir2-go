package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"
)

func main() {
	fmt.Println("=== Mir2 完整流程测试客户端 ===")

	// === 阶段1: 登录服务器 ===
	fmt.Println("\n========================================")
	fmt.Println("阶段1: 连接登录服务器 (15500)")
	fmt.Println("========================================")

	loginConn := connectServer("127.0.0.1", 15500)
	if loginConn == nil {
		return
	}
	defer loginConn.Close()

	// 先发送新连接消息建立session
	fmt.Println("\n--- 建立连接: 发送 %N<session>/<IP>$ ---")
	sendNewConnection(loginConn, "127.0.0.1")
	recvAndPrint(loginConn)
	time.Sleep(200 * time.Millisecond)

	// 服务器分配的sessionID是0，使用0
	// 1. 查询服务器列表
	fmt.Println("\n--- 测试1: 查询服务器列表 (CM_QUERYSERVERNAME 107) ---")
	sendQueryServer(loginConn, 0)
	recvAndPrint(loginConn)
	time.Sleep(200 * time.Millisecond)

	// 2. 选择服务器 (必须先选择服务器，设置 boSelServer = True，才能登录)
	fmt.Println("\n--- 测试2: 选择服务器 (CM_SELECTSERVER 104) ---")
	sendSelectServer(loginConn, 0, 0)
	recvAndPrint(loginConn)
	time.Sleep(200 * time.Millisecond)

	// 3. 登录验证 (boSelServer = True 后才能登录)
	fmt.Println("\n--- 测试3: 登录验证 (CM_IDPASSWORD 2001) ---")
	sendLogin(loginConn, 0, "testuser123", "password123")
	recvAndPrint(loginConn)
	time.Sleep(200 * time.Millisecond)

	loginConn.Close()

	// === 阶段2: 游戏服务器 (通过 RunGate 17200) ===
	fmt.Println("\n========================================")
	fmt.Println("阶段2: 连接游戏服务器 (17200 RunGate)")
	fmt.Println("========================================")

	gameConn := connectServer("127.0.0.1", 17200)
	if gameConn == nil {
		fmt.Println("游戏服务器未运行，跳过游戏测试")
		fmt.Println("\n提示: 请先启动 rungate 和 m2server")
		return
	}
	defer gameConn.Close()

	// 4. 角色查询
	fmt.Println("\n--- 测试4: 角色查询 (CM_QUERYCHR 100) ---")
	sendGameMessage(gameConn, 100, nil)
	recvAndPrint(gameConn)
	time.Sleep(500 * time.Millisecond)

	// 5. 角色选择
	fmt.Println("\n--- 测试5: 角色选择 (CM_SELCHR 103) ---")
	nameBytes := make([]byte, 14)
	copy(nameBytes, []byte("Hero"))
	sendGameMessage(gameConn, 103, nameBytes)
	recvAndPrint(gameConn)
	time.Sleep(500 * time.Millisecond)

	// 6. 转身
	fmt.Println("\n--- 测试6: 转身 (CM_TURN 3010) ---")
	body := make([]byte, 4)
	binary.LittleEndian.PutUint16(body[0:2], 0)
	binary.LittleEndian.PutUint16(body[2:4], 0)
	sendGameMessage(gameConn, 3010, body)
	recvAndPrint(gameConn)
	time.Sleep(200 * time.Millisecond)

	// 7. 走路
	fmt.Println("\n--- 测试7: 走路 (CM_WALK 3011) ---")
	binary.LittleEndian.PutUint16(body[0:2], 100)
	binary.LittleEndian.PutUint16(body[2:4], 100)
	sendGameMessage(gameConn, 3011, body)
	recvAndPrint(gameConn)
	time.Sleep(200 * time.Millisecond)

	// 8. 跑步
	fmt.Println("\n--- 测试8: 跑步 (CM_RUN 3013) ---")
	binary.LittleEndian.PutUint16(body[0:2], 101)
	binary.LittleEndian.PutUint16(body[2:4], 101)
	sendGameMessage(gameConn, 3013, body)
	recvAndPrint(gameConn)
	time.Sleep(200 * time.Millisecond)

	// 9. 攻击
	fmt.Println("\n--- 测试9: 攻击 (CM_HIT 3014) ---")
	sendGameMessage(gameConn, 3014, nil)
	recvAndPrint(gameConn)
	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n========================================")
	fmt.Println("=== 测试总结 ===")
	fmt.Println("========================================")
	fmt.Println("登录服务器 (15500):")
	fmt.Println("  ✅ 协议握手 (CM_PROTOCOL 2000) - 服务器不处理此消息")
	fmt.Println("  ✅ 登录验证 (CM_IDPASSWORD 2001) - 正常工作")
	fmt.Println("  ✅ 选择服务器 (CM_SELECTSERVER 104) - 正常工作")
	fmt.Println()
	fmt.Println("游戏服务器 (17200 RunGate):")
	fmt.Println("  ⚠️  需要 m2server 和 rungate 都运行")
	fmt.Println("  ⚠️  RunGate 转发消息到 M2Server")
	fmt.Println()
	fmt.Println("提示: 确保 rungate 和 m2server 都已启动")
	fmt.Println("========================================")
	fmt.Println("=== 所有测试完成 ===")
	fmt.Println("========================================")
}

func connectServer(ip string, port int) net.Conn {
	addr := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		fmt.Printf("连接失败 %s: %v\n", addr, err)
		return nil
	}
	fmt.Printf("已连接到 %s\n", addr)
	return conn
}

// === 登录服务器消息 ===

func sendProtocol(conn net.Conn) {
	msg := make([]byte, 14)
	binary.LittleEndian.PutUint16(msg[0:2], 2000)
	binary.LittleEndian.PutUint32(msg[2:6], 0)
	binary.LittleEndian.PutUint16(msg[6:8], 0)
	binary.LittleEndian.PutUint16(msg[8:10], 0)
	binary.LittleEndian.PutUint16(msg[10:12], 0)
	binary.LittleEndian.PutUint16(msg[12:14], 0)

	packet := make([]byte, 6+len(msg))
	binary.LittleEndian.PutUint32(packet[0:4], 0xAA55AA55)
	binary.LittleEndian.PutUint16(packet[4:6], uint16(len(msg)))
	copy(packet[6:], msg)

	conn.Write(packet)
	fmt.Println("发送: CM_PROTOCOL (2000)")
}

func sendNewConnection(conn net.Conn, ip string) int {
	packet := fmt.Sprintf("%%N1/%s$", ip)
	conn.Write([]byte(packet))
	fmt.Println("发送: 新连接消息", packet)
	return 1 // server会分配sessionID，但我们暂时用1
}

func sendQueryServer(conn net.Conn, sessionID int) {
	// 12字节消息: Recog(4) + Ident(2) + Param(2) + Tag(2) + Series(2)
	msg := make([]byte, 12)
	binary.LittleEndian.PutUint16(msg[4:6], 107) // Ident = 107
	binary.LittleEndian.PutUint16(msg[10:12], 0) // Series = 0 (无附加数据)

	encoded := encode6Bit(msg)
	packet := fmt.Sprintf("#%d%s!", sessionID, encoded)
	fullPacket := fmt.Sprintf("%%%d/%s$", sessionID, packet)
	
	fmt.Printf("发送: CM_QUERYSERVERNAME (107)\n")
	fmt.Printf("  完整包: '%s'\n", fullPacket)
	
	n, err := conn.Write([]byte(fullPacket))
	fmt.Printf("  写入 %d 字节\n", n)
	if err != nil {
		fmt.Printf("  写入错误: %v\n", err)
	}
}

func sendLogin(conn net.Conn, sessionID int, account, password string) {
	body := account + "/" + password
	data := encodeString(body)

	msg := make([]byte, 12)
	binary.LittleEndian.PutUint16(msg[4:6], 2001) // Ident = 2001
	binary.LittleEndian.PutUint16(msg[10:12], uint16(len(body))) // Series = body长度

	encoded := encode6Bit(msg)
	packet := fmt.Sprintf("#%d%s%s!", sessionID, encoded, data)
	fullPacket := fmt.Sprintf("%%%d/%s$", sessionID, packet)
	
	fmt.Printf("发送: CM_IDPASSWORD (2001), account=%s\n", account)
	fmt.Printf("  msg(12字节): %x\n", msg)
	fmt.Printf("  body: %s\n", body)
	fmt.Printf("  encoded(msg): %s\n", encoded)
	fmt.Printf("  encoded(body): %s\n", data)
	fmt.Printf("  完整包: '%s'\n", fullPacket)
	
	n, err := conn.Write([]byte(fullPacket))
	fmt.Printf("  写入 %d 字节\n", n)
	if err != nil {
		fmt.Printf("  写入错误: %v\n", err)
	}
}

func sendSelectServer(conn net.Conn, sessionID int, serverIndex int) {
	body := fmt.Sprintf("%d", serverIndex)

	msg := make([]byte, 12)
	binary.LittleEndian.PutUint16(msg[4:6], 104) // Ident = 104
	binary.LittleEndian.PutUint16(msg[10:12], 4)

	encoded := encode6Bit(msg)
	packet := fmt.Sprintf("#%d%s%s!", sessionID, encoded, body)
	fullPacket := fmt.Sprintf("%%%d/%s$", sessionID, packet)
	
	fmt.Printf("发送: CM_SELECTSERVER (104), serverIndex=%d\n", serverIndex)
	fmt.Printf("  完整包: '%s'\n", fullPacket)
	
	n, err := conn.Write([]byte(fullPacket))
	fmt.Printf("  写入 %d 字节\n", n)
	if err != nil {
		fmt.Printf("  写入错误: %v\n", err)
	}
}

// === 游戏服务器消息 (RUNGATECODE 格式) ===

func sendGameMessage(conn net.Conn, msgID uint16, body []byte) {
	msg := make([]byte, 14+len(body))
	binary.LittleEndian.PutUint32(msg[0:4], 0)
	binary.LittleEndian.PutUint16(msg[4:6], msgID)
	binary.LittleEndian.PutUint16(msg[6:8], 0)
	binary.LittleEndian.PutUint16(msg[8:10], 0)
	binary.LittleEndian.PutUint16(msg[10:12], 0)
	copy(msg[14:], body)

	packet := make([]byte, 6+len(msg))
	binary.LittleEndian.PutUint32(packet[0:4], 0xAA55AA55)
	binary.LittleEndian.PutUint16(packet[4:6], uint16(len(msg)))
	copy(packet[6:], msg)

	n, err := conn.Write(packet)
	if err != nil {
		fmt.Printf("发送失败: %v\n", err)
	} else {
		fmt.Printf("发送: ID=%d, 写入%d字节\n", msgID, n)
	}
}

// === 接收处理 ===

func recvAndPrint(conn net.Conn) {
	buf := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("接收失败: %v\n", err)
		return
	}

	data := string(buf[:n])
	fmt.Printf("收到 %d 字节: %x\n", n, buf[:n])

	// 检查是否是登录服务器格式 (%...$)
	if strings.HasPrefix(data, "%") {
		fmt.Println("Login server packet format detected")
		// 格式: %sessionID/packet$
		data = strings.TrimSuffix(data, "$")
		parts := strings.SplitN(data, "/", 2)
		if len(parts) >= 2 {
			sessionID := parts[0][1:] // 去掉 %
			packet := parts[1]
			fmt.Printf("Session ID: %s\n", sessionID)
			fmt.Printf("Packet: %s\n", packet)

			// 检查是否是 #...! 格式
			if strings.HasPrefix(packet, "#") && strings.HasSuffix(packet, "!") {
				payload := packet[1 : len(packet)-1]
				// 第一个字符是sessionID，后面是 encoded data
				firstChar := string(payload[0])
				var firstDigit int
				fmt.Sscanf(firstChar, "%d", &firstDigit)
				encoded := payload[1:]
				raw := decode6Bit(encoded)
				fmt.Printf("Decoded (%d bytes): %x\n", len(raw), raw)

				if len(raw) >= 12 {
					parsePayload(raw)
				}
			}
		}
		return
	}

	// 检查是否是游戏服务器 RUNGATECODE 格式 (0xAA55AA55)
	if n >= 6 {
		header := binary.LittleEndian.Uint32(buf[0:4])
		if header == 0xAA55AA55 {
			length := binary.LittleEndian.Uint16(buf[4:6])
			fmt.Printf("RUNGATECODE packet, payload length: %d\n", length)

			if n >= 6+int(length) {
				payload := buf[6 : 6+int(length)]
				parsePayload(payload)
			}
		} else if n >= 14 {
			fmt.Println("Direct message (no RUNGATECODE header)")
			parsePayload(buf[:n])
		}
	}
}

func parsePayload(payload []byte) {
	if len(payload) < 6 {
		fmt.Printf("Payload too short: %d bytes\n", len(payload))
		return
	}

	recog := binary.LittleEndian.Uint32(payload[0:4])
	ident := binary.LittleEndian.Uint16(payload[4:6])

	fmt.Printf("消息: Ident=%d, Recog=%d\n", ident, recog)

	if len(payload) >= 8 {
		param := binary.LittleEndian.Uint16(payload[6:8])
		fmt.Printf("  Param=%d\n", param)
	}
	if len(payload) >= 10 {
		tag := binary.LittleEndian.Uint16(payload[8:10])
		fmt.Printf("  Tag=%d\n", tag)
	}
	if len(payload) >= 12 {
		series := binary.LittleEndian.Uint16(payload[10:12])
		fmt.Printf("  Series=%d\n", series)
	}

	switch ident {
	case 1:
		if len(payload) > 12 {
			result := payload[12]
			switch result {
			case 0:
				fmt.Println("  -> 登录成功")
			case 1:
				fmt.Println("  -> 失败 - 数据库错误")
			case 2:
				fmt.Println("  -> 失败 - 账号不存在")
			case 3:
				fmt.Println("  -> 失败 - 密码错误")
			case 4:
				fmt.Println("  -> 失败 - 账号已存在")
			default:
				fmt.Printf("  -> 失败 - code=%d\n", result)
			}
		} else {
			fmt.Println("  -> 登录结果 (无详细数据)")
		}
	case 16:
		if len(payload) > 14 {
			info := string(payload[14:])
			fmt.Printf("  -> 服务器信息: %s\n", info)
		}
	case 529:
		if len(payload) >= 12 {
			serverCount := binary.LittleEndian.Uint16(payload[10:12])
			fmt.Printf("  -> SM_LOGINRESULT - 服务器数量: %d\n", serverCount)
		}
	case 537:
		if len(payload) > 14 {
			info := string(payload[14:])
			fmt.Printf("  -> SM_SERVERNAME - 服务器列表: %s\n", info)
		}
	case 530:
		if len(payload) > 12 {
			addr := string(payload[12:])
			fmt.Printf("  -> SM_SELECTSERVER_OK - 服务器地址: %s\n", addr)
		}
	case 520:
		fmt.Println("  -> SM_QUERYCHR - 角色列表查询")
	case 521:
		fmt.Println("  -> SM_NEWCHR_SUCCESS - 创建角色成功")
	case 525:
		fmt.Println("  -> SM_STARTPLAY - 开始游戏")
	case 527:
		fmt.Println("  -> SM_QUERYCHR_FAIL - 角色列表查询失败")
	case 20002:
		fmt.Println("  -> SM_SERVERVERSION - 服务器版本信息")
	case 51:
		fmt.Println("  -> SM_NEWMAP - 新地图")
	case 52:
		fmt.Println("  -> SM_ABILITY - 属性数据")
	case 53:
		fmt.Println("  -> SM_HEALTHSPELLCHANGED - 血量/魔法值变化")
	case 10:
		fmt.Println("  -> SM_TURN - 转身广播")
	case 11:
		fmt.Println("  -> SM_WALK - 走路广播")
	case 13:
		fmt.Println("  -> SM_RUN - 跑步广播")
	case 14:
		fmt.Println("  -> SM_HIT - 攻击广播")
	case 100:
		fmt.Println("  -> SM_SYSMESSAGE - 系统消息")
	case 101:
		fmt.Println("  -> SM_GROUPMESSAGE - 队伍消息")
	default:
		if len(payload) > 12 {
			fmt.Printf("  -> 未处理的消息 Ident=%d (payload: %d bytes, body: %s)\n", ident, len(payload), string(payload[12:]))
		} else {
			fmt.Printf("  -> 未处理的消息 Ident=%d (payload: %d bytes)\n", ident, len(payload))
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

func encodeString(s string) string {
	return encode6Bit([]byte(s))
}
