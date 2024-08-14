package main

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/pedrogomes29/blockchain/blockchain"
)

const genesisAddress = "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

func main() {
	bc := blockchain.NewBlockchain(genesisAddress)
	balance := 0
	pubKeyHash, _, _ := base58.CheckDecode(genesisAddress)
	UTXOs, _ := bc.FindUTXOs(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", genesisAddress, balance)
}
