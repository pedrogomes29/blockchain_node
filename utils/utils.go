package utils

import (
	"encoding/binary"
)

func Uint32ToHex(num uint32) []byte {
	hex := make([]byte, 4)
	binary.LittleEndian.PutUint32(hex, num)
	return hex
}
