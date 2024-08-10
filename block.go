package main

import (
	"github.com/pedrogomes29/blockchain/transactions"
)

type BlockHeader struct {
	PrevBlocHeaderkHash []byte
	MerkleRootHash      []byte
	Nonce               uint32
}

type Block struct {
	Header       BlockHeader
	Transactions []*transactions.Transaction
}