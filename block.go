package main

import (
	"bytes"
	"crypto/sha256"
	"math"
	"math/big"

	"github.com/pedrogomes29/blockchain/merkle_tree"
	"github.com/pedrogomes29/blockchain/transactions"
	"github.com/pedrogomes29/blockchain/utils"
)

type BlockHeader struct {
	PrevBlockHeaderHash []byte
	MerkleRootHash      []byte
	Nonce               uint32
}

type Block struct {
	Header       BlockHeader
	Transactions []*transactions.Transaction
}

const MaxNonce = math.MaxUint32
const targetBits = 8 //how many bits must be 0 in the header hash

var Target *big.Int

func init() {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))
}

func NewBlock(transactions []*transactions.Transaction, prevBlockHash []byte) *Block {
	blockHeader := BlockHeader{
		PrevBlockHeaderHash: prevBlockHash,
	}

	block := &Block{
		Header:       blockHeader,
		Transactions: transactions,
	}

	return block
}

func (b *BlockHeader) getBlockHeaderHash() [32]byte {
	data := bytes.Join(
		[][]byte{
			b.PrevBlockHeaderHash,
			b.MerkleRootHash,
		},
		utils.Uint32ToHex(b.Nonce),
	)

	return sha256.Sum256(data)
}

func (bh *BlockHeader) generateNoncePOW() {
	for possibleNonce := 0; possibleNonce < MaxNonce; possibleNonce++ {
		bh.Nonce = uint32(possibleNonce)
		if bh.ValidateNonce() {
			break;
		}
	}
}

func (b *BlockHeader) ValidateNonce() bool {
	var hashInt big.Int

	hash := b.getBlockHeaderHash()
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(Target) == -1

	return isValid
}

func (b *Block) generateMerkleRootHash() {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}

	mTree := merkle_tree.NewMerkleTree(transactions)

	b.Header.MerkleRootHash = mTree.RootNode.Data
}
