package blockchain

import (
	"fmt"
	"log"

	"github.com/pedrogomes29/blockchain_node/transactions"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

type Blockchain struct {
	LastBlockHash []byte
	BlocksDB      *leveldb.DB
	ChainstateDB  *leveldb.DB
}

func (bc *Blockchain) AddBlock(newBlock *Block) {
	blockHash := newBlock.GetBlockHeaderHash()
	bc.LastBlockHash = blockHash[:]
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
	var LastBlockHash []byte
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
		LastBlockHash = genesisHash[:]

		err = blocksDB.Put(LastBlockHash, genesis.Serialize(), nil)
		if err != nil {
			log.Panic(err)
		}
		err = blocksDB.Put([]byte("l"), LastBlockHash, nil)
		if err != nil {
			log.Panic(err)
		}
		bc = &Blockchain{LastBlockHash, blocksDB, chainstateDB}
		bc.ReindexUTXOs()
	} else if err != nil {
		log.Panic(err)
	} else { //else, simply get the last block hash from the db
		fmt.Println("Blockchain found. Retrieving...")
		LastBlockHash, err = blocksDB.Get([]byte("l"), nil)

		if err != nil {
			log.Panic(err)
		}

		bc = &Blockchain{LastBlockHash, blocksDB, chainstateDB}
	}

	return bc
}
