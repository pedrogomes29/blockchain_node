package utils

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
)

func Uint32ToHex(num uint32) []byte {
	hex := make([]byte, 4)
	binary.LittleEndian.PutUint32(hex, num)
	return hex
}


func GenerateRandomString(nrBytes int) string{
	randData := make([]byte, nrBytes)
	_, err := rand.Read(randData)
	if err != nil {
		log.Panic(err)
	}

	return fmt.Sprintf("%x", randData)
}