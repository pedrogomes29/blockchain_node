package main

import (
	"encoding/hex"
	"fmt"

	"github.com/pedrogomes29/blockchain/blockchain"
)

const genesisAddress = "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

func main() {
	bc := blockchain.NewBlockchain(genesisAddress)
	fmt.Println(hex.EncodeToString(bc.LastBlockHash))
}
