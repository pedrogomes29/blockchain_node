package server

import (
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
	id     commandID
	peer   *peer
	args   [][]byte
}


type versionPayload struct {
	BestHeight int
	ACK bool //whether an acknowledgement is piggybacked
}
func ParseVersionPayload (args [][]byte) versionPayload{
	bestHeight,err := strconv.Atoi(string(args[0]))
	if err!=nil{
		log.Panicf("error parsing peer's blockchain height %s", string(args[0]))
	}
	versionPayload := versionPayload{
		BestHeight: bestHeight,
	}
	if len(args)>1 && string(args[1])=="ACK"{
		versionPayload.ACK = true
	}
	return versionPayload
}


type addrPayload []string
func ParseAddrsPayload(args [][]byte) addrPayload{
    payload := make([]string, len(args))
    for i, arg := range args {
        payload[i] = string(arg)
    }
    return addrPayload(payload)
}

type blockHeaderHash []byte
type getBlocksPayload []blockHeaderHash
func ParseGetBlocksPayload(args [][]byte) getBlocksPayload{
    payload := make(getBlocksPayload, len(args))
    for i, arg := range args {
        payload[i] = blockHeaderHash(arg)
    }
    return payload
}


type objectEntry struct{
	objectType objectType
	object []byte
}

func ParseObjects(args [][]byte) []objectEntry{
    payload := make([]objectEntry, len(args)/2)
	for i:=0;i<len(args);i+=2{
		entry := objectEntry{}
		entryTypeString := string(args[i])
		switch entryTypeString {
		case "MSG_TX":
			entry.objectType = TX
		case "MSG_BLOCK":
			entry.objectType = BLOCK
		}
		entry.object = args[i+1]
		payload = append(payload,entry)
	}
    return payload
}