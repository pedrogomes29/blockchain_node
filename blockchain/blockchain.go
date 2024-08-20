package blockchain

import (
	"fmt"
	"log"
	"slices"

	"github.com/pedrogomes29/blockchain_node/transactions"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

type Blockchain struct {
	LastBlockHash []byte
	Height        int
	BlocksDB      *leveldb.DB
	ChainstateDB  *leveldb.DB
}

func (bc *Blockchain) AddBlock(newBlock *Block) {
	blockHash := newBlock.GetBlockHeaderHash()
	bc.LastBlockHash = blockHash[:]
	bc.Height++
	err := bc.BlocksDB.Put(bc.LastBlockHash, newBlock.Serialize(), nil)
	if err != nil {
		log.Panic(err)
	}

	for _, tx := range newBlock.Transactions {
		err := tx.IndexUTXOs(bc.ChainstateDB)
		if err != nil {
			log.Panic(err)
		}
	}
}

func NewBlockchain(genesisAddress string) *Blockchain {
	var lastBlockHash []byte
	var bc *Blockchain

	blocksDB, err := leveldb.OpenFile("blocks", nil)
	if err != nil {
		log.Panic(err)
	}

	chainstateDB, err := leveldb.OpenFile("chainstate", nil)
	if err != nil {
		log.Panic(err)
	}

	_, err = blocksDB.Get([]byte("l"), nil)

	if err == errors.ErrNotFound { //if the l key (last block hash) is not found, we are creating the db for the first time => create genesis block
		fmt.Println("Blockchain not found. Generating genesis block...")
		cbtx := transactions.NewCoinbaseTX(genesisAddress)
		genesis := NewGenesisBlock(cbtx)
		genesisHash := genesis.GetBlockHeaderHash()
		lastBlockHash = genesisHash[:]

		err = blocksDB.Put(lastBlockHash, genesis.Serialize(), nil)
		if err != nil {
			log.Panic(err)
		}
		err = blocksDB.Put([]byte("l"), lastBlockHash, nil)
		if err != nil {
			log.Panic(err)
		}
		bc = &Blockchain{lastBlockHash, 1, blocksDB, chainstateDB}
		bc.ReindexUTXOs()
	} else if err != nil {
		log.Panic(err)
	} else { //else, simply get the last block hash from the db
		fmt.Println("Blockchain found. Retrieving...")
		lastBlockHash, err = blocksDB.Get([]byte("l"), nil)
		if err != nil {
			log.Panic(err)
		}
		blockBytes, err := blocksDB.Get(lastBlockHash, nil)
		if err != nil {
			log.Panic(err)
		}
		lastBlock := DeserializeBlock(blockBytes)

		fmt.Printf("Retrieved blockchain with height:%d\n", lastBlock.Header.Height)

		bc = &Blockchain{lastBlockHash, lastBlock.Header.Height, blocksDB, chainstateDB}
	}

	return bc
}


func (bc *Blockchain) getBlocks() []*Block {
	prevBlockHash := bc.LastBlockHash
	var blocks []*Block 

	for len(prevBlockHash) > 0 {
		blockBytes, err := bc.BlocksDB.Get(prevBlockHash, nil)
		if err != nil {
			log.Panic(err)
		}
		block := DeserializeBlock(blockBytes)
		blocks = append(blocks,block)
		prevBlockHash = block.Header.PrevBlockHeaderHash
	}

	slices.Reverse(blocks)
	return blocks
}


func (bc *Blockchain) GetLast5BlockHeaders() [][]byte{
	var blockHashes [][]byte
	prevBlockHash := bc.LastBlockHash
	for i:=0; i<5 && len(prevBlockHash) > 0; i++{
		blockBytes, err := bc.BlocksDB.Get(prevBlockHash, nil)
		if err != nil {
			log.Panic(err)
		}
		block := DeserializeBlock(blockBytes)
		blockHash := block.GetBlockHeaderHash()
		blockHashes = append(blockHashes, blockHash[:])
	}

	return blockHashes
}