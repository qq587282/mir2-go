package utils

import (
	"crypto/md5"
	"encoding/binary"
	"encoding/hex"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func MD5(data string) string {
	hash := md5.Sum([]byte(data))
	return hex.EncodeToString(hash[:])
}

func MD5Bytes(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

func RandomInt(min, max int) int {
	if min >= max {
		return min
	}
	return rand.Intn(max-min) + min
}

func RandomInt64(min, max int64) int64 {
	if min >= max {
		return min
	}
	return rand.Int63n(max-min) + min
}

func RandomBytes(length int) []byte {
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = byte(rand.Intn(256))
	}
	return result
}

func IntToBytes(value int32, buf []byte) {
	binary.LittleEndian.PutUint32(buf, uint32(value))
}

func BytesToInt(buf []byte) int32 {
	return int32(binary.LittleEndian.Uint32(buf))
}

func Int16ToBytes(value int16, buf []byte) {
	binary.LittleEndian.PutUint16(buf, uint16(value))
}

func BytesToInt16(buf []byte) int16 {
	return int16(binary.LittleEndian.Uint16(buf))
}

func UIntToBytes(value uint32, buf []byte) {
	binary.LittleEndian.PutUint32(buf, value)
}

func BytesToUInt(buf []byte) uint32 {
	return binary.LittleEndian.Uint32(buf)
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

func AbsInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func InRange(x1, y1, x2, y2, range_ int) bool {
	dx := x1 - x2
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y2
	if dy < 0 {
		dy = -dy
	}
	return dx <= range_ && dy <= range_
}

func Distance(x1, y1, x2, y2 int) int {
	dx := x1 - x2
	if dx < 0 {
		dx = -dx
	}
	dy := y1 - y2
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}

func DistanceFloat(x1, y1, x2, y2 float64) float64 {
	dx := x2 - x1
	dy := y2 - y1
	return dx*dx + dy*dy
}

func GetDirection(x1, y1, x2, y2 int) int {
	dx := x2 - x1
	dy := y2 - y1

	if dx == 0 && dy < 0 {
		return 0
	}
	if dx > 0 && dy < 0 {
		return 1
	}
	if dx > 0 && dy == 0 {
		return 2
	}
	if dx > 0 && dy > 0 {
		return 3
	}
	if dx == 0 && dy > 0 {
		return 4
	}
	if dx < 0 && dy > 0 {
		return 5
	}
	if dx < 0 && dy == 0 {
		return 6
	}
	if dx < 0 && dy < 0 {
		return 7
	}

	return 0
}

func TrimString(s []byte) string {
	end := 0
	for i, b := range s {
		if b == 0 {
			end = i
			break
		}
		end = i + 1
	}
	return string(s[:end])
}

func PadString(s string, length int) []byte {
	result := make([]byte, length)
	copy(result, []byte(s))
	return result
}

func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func IntToBool(i int) bool {
	return i != 0
}

func PercentChance(percent int) bool {
	if percent <= 0 {
		return false
	}
	if percent >= 100 {
		return true
	}
	return rand.Intn(100) < percent
}

func CompareStringIgnoreCase(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		ca := a[i]
		cb := b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}
