package blockchain

import (
	"fmt"
	"log"

	"github.com/pedrogomes29/blockchain/transactions"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

type Blockchain struct {
	LastBlockHash []byte
	blocksDB  *leveldb.DB
	chainstateDB  *leveldb.DB
}

func (bc *Blockchain) AddBlock(transactions []*transactions.Transaction) {
	newBlock := NewBlock(transactions, bc.LastBlockHash)
	newBlock.GenerateMerkleRootHash()
	newBlock.Header.GenerateNoncePOW()
	blockHash := newBlock.Header.GetBlockHeaderHash()
	bc.LastBlockHash = blockHash[:]
	err := bc.blocksDB.Put(bc.LastBlockHash, newBlock.Serialize(), nil)
	if err != nil {
		log.Panic(err)
	}
}

func NewBlockchain(genesisAddress string) *Blockchain {
	var LastBlockHash []byte
	blocksDB, err := leveldb.OpenFile("blocks", nil)
	if err != nil {
		log.Panic(err)
	}
	_, err = blocksDB.Get([]byte("l"), nil)

	if err == errors.ErrNotFound { //if the l key (last block hash) is not found, we are creating the db for the first time => create genesis block
		fmt.Println("Blockchain not found. Generating genesis block...")
		cbtx := transactions.NewCoinbaseTX(genesisAddress)
		genesis := NewGenesisBlock(cbtx)
		genesisHash := genesis.Header.GetBlockHeaderHash()
		LastBlockHash = genesisHash[:]

		err = blocksDB.Put(LastBlockHash, genesis.Serialize(), nil)
		if err != nil {
			log.Panic(err)
		}
		err = blocksDB.Put([]byte("l"), LastBlockHash, nil)
		if err != nil {
			log.Panic(err)
		}

	} else { //else, simply get the last block hash from the db
		fmt.Println("Blockchain found. Retrieving...")
		LastBlockHash, err = blocksDB.Get([]byte("l"),nil)
		if err != nil {
			log.Panic(err)
		}
	}

	return &Blockchain{LastBlockHash,blocksDB,nil}
}
