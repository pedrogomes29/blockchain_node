package main

import "github.com/pedrogomes29/blockchain/transactions"

const genesisAddress = "1A1zP1eP5QGefi2DMPTfTL5SLmv7DivfNa"

type Blockchain struct {
	blocks []*Block
}

func (bc *Blockchain) AddBlock(transactions []*transactions.Transaction) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	prevBlockHash := prevBlock.Header.GetBlockHeaderHash()
	newBlock := NewBlock(transactions, prevBlockHash[:])
	bc.blocks = append(bc.blocks, newBlock)
}

func NewBlockchain() *Blockchain {
	cbtx := transactions.NewCoinbaseTX(genesisAddress)
	genesis := NewGenesisBlock(cbtx)
	genesis.GenerateMerkleRootHash()
	genesis.Header.GenerateNoncePOW()
	return &Blockchain{[]*Block{genesis}}
}