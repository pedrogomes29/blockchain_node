package main

import (
	"encoding/hex"
	"fmt"
)

func main() {
	bc := NewBlockchain()
	genesisBlock := bc.blocks[0]
	genesisBlockHeader := genesisBlock.Header
	genesisBlockNonce := genesisBlockHeader.Nonce
	genesisBlockHash := genesisBlockHeader.GetBlockHeaderHash()

	fmt.Println(genesisBlockNonce)
	fmt.Println(hex.EncodeToString(genesisBlockHash[:]))
}
