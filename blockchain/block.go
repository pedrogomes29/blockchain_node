package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"math"
	"math/big"
	"time"

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
	Target = big.NewInt(1)
	Target.Lsh(Target, uint(256-targetBits))
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

func NewGenesisBlock(coinbase *transactions.Transaction) *Block {
	genesisblock := NewBlock([]*transactions.Transaction{coinbase}, []byte{})
	done := make(chan struct{})
	go genesisblock.POW(done)
	<-done
	return genesisblock
}

func (b *Block) GetBlockHeaderHash() [32]byte {
	b.GenerateMerkleRootHash()
	data := bytes.Join(
		[][]byte{
			b.Header.PrevBlockHeaderHash,
			b.Header.MerkleRootHash,
		},
		utils.Uint32ToHex(b.Header.Nonce),
	)

	return sha256.Sum256(data)
}

func (b *Block) POW(done chan struct{}) {
	for possibleNonce := 0; possibleNonce < MaxNonce; possibleNonce++ {
		b.Header.Nonce = uint32(possibleNonce)
		if b.ValidateNonce() {
			break
		}
		time.Sleep(time.Microsecond*50)
	}
	done <- struct{}{}
}

func (b *Block) ValidateNonce() bool {
	var hashInt big.Int

	hash := b.GetBlockHeaderHash()
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(Target) == -1

	return isValid
}

func (b *Block) GenerateMerkleRootHash() {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}

	mTree := merkle_tree.NewMerkleTree(transactions)

	b.Header.MerkleRootHash = mTree.RootNode.Data
}


func (b *Block) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}