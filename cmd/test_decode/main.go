package main

import (
	"encoding/binary"
	"fmt"
)

func main() {
	// 发送的消息
	sent := "#1<<<<<=`><<<<<<@<PrQnYbQnHNxl!"
	_ = sent
	fmt.Println("发送:", sent)
	
	// 解析
	code := int(sent[1] - '0')
	_ = code
	encoded := sent[2:]
	
	raw := decode6Bit(encoded)
	fmt.Printf("解码 (%d bytes): %x\n", len(raw), raw)
	
	if len(raw) >= 12 {
		recog := binary.LittleEndian.Uint32(raw[0:4])
		ident := binary.LittleEndian.Uint16(raw[4:6])
		param := binary.LittleEndian.Uint16(raw[6:8])
		tag := binary.LittleEndian.Uint16(raw[8:10])
		series := binary.LittleEndian.Uint16(raw[10:12])
		
		fmt.Printf("Header: recog=%d, ident=%d, param=%d, tag=%d, series=%d\n", 
			recog, ident, param, tag, series)
		
		if len(raw) > 12 {
			body := string(raw[12:])
			fmt.Printf("Body: '%s'\n", body)
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