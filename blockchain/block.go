package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
	"math"
	"math/big"
	"time"

	"github.com/pedrogomes29/blockchain_node/merkle_tree"
	"github.com/pedrogomes29/blockchain_node/transactions"
	"github.com/pedrogomes29/blockchain_node/utils"
)

type BlockHeader struct {
	PrevBlockHeaderHash []byte
	MerkleRootHash      []byte
	Nonce               uint32
	Height              int
}

type Block struct {
	Header       BlockHeader
	Transactions []*transactions.Transaction
}

const MaxNonce = math.MaxUint32
const targetBits = 14            //how many bits must be 0 in the header hash
const maxBlockSize = 1024 * 1024 //1 MB

var Target *big.Int

func init() {
	Target = big.NewInt(1)
	Target.Lsh(Target, uint(256-targetBits))
}

func NewBlock(transactions []*transactions.Transaction, prevBlockHash []byte, height int) *Block {
	blockHeader := BlockHeader{
		PrevBlockHeaderHash: prevBlockHash,
		Height:              height,
	}

	block := &Block{
		Header:       blockHeader,
		Transactions: transactions,
	}

	block.Header.MerkleRootHash = block.MerkleRootHash()

	return block
}

func NewGenesisBlock(miningChan chan struct{}, coinbase *transactions.Transaction) *Block {
	genesisblock := NewBlock([]*transactions.Transaction{coinbase}, []byte{}, 0)
	genesisblock.POW(miningChan)
	return genesisblock
}

func (b *Block) GetBlockHeaderHash() [32]byte {
	data := bytes.Join(
		[][]byte{
			b.Header.PrevBlockHeaderHash,
			b.Header.MerkleRootHash,
		},
		utils.Uint32ToHex(b.Header.Nonce),
	)

	return sha256.Sum256(data)
}

func (b *Block) POW(miningChan chan struct{}) bool{
	for possibleNonce := 0; possibleNonce < MaxNonce; possibleNonce++ {
		select {
		case <-miningChan:
			fmt.Printf("Mining interrupted for block %d\n", b.Header.Height)
			return false
		default:
			b.Header.Nonce = uint32(possibleNonce)
			if b.ValidateNonce() {
				fmt.Printf("Mined block %d\n", b.Header.Height)
				return true
			}
			time.Sleep(time.Microsecond * 50)
		}
	}
	return false;
}

func (b *Block) ValidateNonce() bool {
	var hashInt big.Int

	hash := b.GetBlockHeaderHash()
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(Target) == -1

	return isValid
}

func (b *Block) AddTransaction(transaction *transactions.Transaction) bool {
	blockWithTxSize := len(transaction.Serialize()) + len(b.Serialize())
	if blockWithTxSize > maxBlockSize {
		return false
	}
	b.Transactions = append(b.Transactions, transaction)
	b.Header.MerkleRootHash = b.MerkleRootHash()
	return true
}

func (b *Block) MerkleRootHash() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}

	mTree := merkle_tree.NewMerkleTree(transactions)

	return mTree.RootNode.Data
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
