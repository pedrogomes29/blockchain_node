package blockchain

import (
	"bytes"
	"fmt"
	"log"
	"slices"

	"github.com/pedrogomes29/blockchain_node/memory_pool"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

type Blockchain struct {
	BlocksDB     *leveldb.DB
	ChainstateDB *leveldb.DB
}

func (bc *Blockchain) VerifyBlock(block *Block) error {
	if bc.Height() != block.Header.Height-1 {
		return errors.New("new block's height isn't current blockchain height plus 1")
	}
	if !bytes.Equal(block.Header.PrevBlockHeaderHash, bc.LastBlockHash()) {
		return errors.New("received block isn't sucessor of blockchain's last block")
	}
	if !bytes.Equal(block.MerkleRootHash(), block.Header.MerkleRootHash) {
		return errors.New("merkle root doesn't match with transactions")
	}
	if !block.ValidateNonce() {
		return errors.New("nonce isn't valid")
	}
	if err := bc.VerifyBlockTxs(block); err != nil {
		return err
	}

	return nil
}

func (bc *Blockchain) VerifyBlockTxs(block *Block) error {
	memoryPool := memory_pool.NewMemoryPool()

	for _, tx := range block.Transactions {
		//returns an error if signature is invalid or the UTXOs are invalid according to the blockchain state
		//(excluding other transactions in the new block)
		err := tx.Verify(bc.ChainstateDB)
		if err != nil {
			return err
		}

		//returns an error if this tx's UTXOs are already spent by some other tx in the new block
		err = memoryPool.PushBackTxWithLock(tx)
		if err != nil {
			return err
		}

	}
	return nil
}

func (bc *Blockchain) AddBlock(newBlock *Block) error {
	blockHash := newBlock.GetBlockHeaderHash()
	//fmt.Printf("Adding block with hash:%s\n", hex.EncodeToString(blockHash))

	err := bc.VerifyBlock(newBlock)
	if err != nil {
		return err
	}

	err = bc.BlocksDB.Put(blockHash, newBlock.Serialize(), nil)
	if err != nil {
		return err
	}

	err = bc.BlocksDB.Put([]byte("l"), blockHash, nil)
	if err != nil {
		return err
	}

	for _, tx := range newBlock.Transactions {
		err = tx.IndexUTXOs(bc.ChainstateDB)
		if err != nil {
			return err
		}
	}

	return nil
}

func (bc *Blockchain) RemoveBlock(blockHash []byte) error {
	//fmt.Printf("Removing block with hash:%s\n", hex.EncodeToString(blockHash))
	if !bytes.Equal(blockHash, bc.LastBlockHash()) {
		return errors.New("can only remove last block")
	}

	block := bc.GetBlock(blockHash)

	if block == nil {
		return errors.New("block not found")
	}

	err := bc.BlocksDB.Delete(blockHash, nil) //delete block from db
	if err != nil {
		return err
	}

	err = bc.BlocksDB.Put([]byte("l"), block.Header.PrevBlockHeaderHash, nil) //update last block hash in db
	if err != nil {
		return err
	}

	for _, tx := range block.Transactions {
		err := tx.RevertUTXOIndex(bc.ChainstateDB)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewBlockchain(miningChan chan struct{}, genesisAddress string) *Blockchain {
	blocksDB, err := leveldb.OpenFile("blocks", nil)
	if err != nil {
		log.Panic(err)
	}

	chainstateDB, err := leveldb.OpenFile("chainstate", nil)
	if err != nil {
		log.Panic(err)
	}

	_, err = blocksDB.Get([]byte("l"), nil)

	if err == leveldb.ErrNotFound {
		//if the l key (last block hash) is not found, we are creating the db for the first time => set "l" as empty []byte
		fmt.Println("Blockchain not found. Creating...")
		err = blocksDB.Put([]byte("l"), []byte{}, nil)
		if err != nil {
			log.Panic(err)
		}
	} else if err != nil {
		log.Panic(err)
	} else {
		fmt.Println("Blockchain found. Retrieving...")
	}

	return &Blockchain{blocksDB, chainstateDB}
}

// gets blocks from older to more recent starting from (but excluding) the argument received in the argument
func (bc *Blockchain) GetBlocksStartingAtHash(hash []byte) []*Block {
	var blocks []*Block
	blockHash := bc.LastBlockHash()
	for !bytes.Equal(blockHash, hash) {
		blockBytes, err := bc.BlocksDB.Get(blockHash, nil)
		if err != nil {
			log.Panic(err)
		}
		block := DeserializeBlock(blockBytes)
		blocks = append(blocks, block)
		blockHash = block.Header.PrevBlockHeaderHash
	}

	slices.Reverse(blocks)
	return blocks
}

func (bc *Blockchain) GetLastBlockHashes(nrHashes int) [][]byte {
	var blockHashes [][]byte
	blockHash := bc.LastBlockHash()
	for i := 0; i < nrHashes && len(blockHash) > 0; i++ {
		blockHashes = append(blockHashes, blockHash[:])
		blockBytes, err := bc.BlocksDB.Get(blockHash, nil)
		if err != nil {
			log.Panic(err)
		}
		block := DeserializeBlock(blockBytes)
		blockHash = block.Header.PrevBlockHeaderHash
	}

	return blockHashes
}

func (bc *Blockchain) GetBlock(blockHash []byte) *Block {
	blockBytes, err := bc.BlocksDB.Get(blockHash, nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			return nil
		}
		log.Panic(err)
	}
	block := DeserializeBlock(blockBytes)
	return block
}

func (bc *Blockchain) Height() int {
	lastBlockHash, err := bc.BlocksDB.Get([]byte("l"), nil)
	if err != nil {
		log.Panic(err)
	}

	if bytes.Equal(lastBlockHash, []byte{}) { //no genesis block
		return -1
	}

	blockBytes, err := bc.BlocksDB.Get(lastBlockHash, nil)
	if err != nil {
		log.Panic(err)
	}
	lastBlock := DeserializeBlock(blockBytes)
	return lastBlock.Header.Height
}

func (bc *Blockchain) LastBlockHash() []byte {
	lastBlockHash, err := bc.BlocksDB.Get([]byte("l"), nil)
	if err != nil {
		log.Panic(err)
	}
	return lastBlockHash
}
