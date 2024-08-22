package server

import (
	"encoding/hex"
	"log"
	"strconv"
)

type commandID int

const (
	GET_ADDR commandID = iota
	ADDR
	VERSION
	VERSION_ACK
	GET_BLOCKS
	INV
	GET_DATA
	DATA
)

type objectType int

const (
	TX objectType = iota
	BLOCK
)

type command struct {
	id   commandID
	peer *peer
	args []string
}

type versionPayload struct {
	BestHeight int
	ACK        bool //whether an acknowledgement is piggybacked
}

func ParseVersionPayload(args []string) versionPayload {
	bestHeight, err := strconv.Atoi(args[0])
	if err != nil {
		log.Panicf("error parsing peer's blockchain height %s", args[0])
	}
	versionPayload := versionPayload{
		BestHeight: bestHeight,
	}
	if len(args) > 1 && args[1] == "ACK" {
		versionPayload.ACK = true
	}
	return versionPayload
}

type addrPayload []string

func ParseAddrsPayload(args []string) addrPayload {
	return addrPayload(args)
}

type blockHeaderHash []byte
type getBlocksPayload []blockHeaderHash

func ParseGetBlocksPayload(args []string) getBlocksPayload {
	payload := make(getBlocksPayload, len(args))
	for i, arg := range args {
		payload[i], _ = hex.DecodeString(arg) //TODO: Error handling
	}
	return payload
}

type objectEntries struct {
	blockEntries [][]byte
	txEntries    [][]byte
}

func ParseObjects(args []string) objectEntries {
	var blockEntries [][]byte
	var txEntries [][]byte
	for i := 0; i < len(args); i += 2 {
		entryTypeString := args[i]
		entry, _ := hex.DecodeString(args[i+1]) //TODO: Error handling

		switch entryTypeString {
		case "TX":
			txEntries = append(txEntries, entry)
		case "BLOCK":
			blockEntries = append(blockEntries, entry)
		}
	}
	return objectEntries{
		blockEntries: blockEntries,
		txEntries:    txEntries,
	}
}
