package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"

	"golang.org/x/crypto/ripemd160"
)

func Uint32ToHex(num uint32) []byte {
	hex := make([]byte, 4)
	binary.LittleEndian.PutUint32(hex, num)
	return hex
}

func GenerateRandomString(nrBytes int) string {
	randData := make([]byte, nrBytes)
	_, err := rand.Read(randData)
	if err != nil {
		log.Panic(err)
	}

	return fmt.Sprintf("%x", randData)
}

func HashPublicKey(publicKey []byte) []byte {
	publicSHA256 := sha256.Sum256(publicKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}
