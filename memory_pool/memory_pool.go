package memory_pool

import (
	"container/list"
	"encoding/hex"
	"sync"

	"github.com/pedrogomes29/blockchain_node/blockchain"
	"github.com/pedrogomes29/blockchain_node/transactions"
)

type MemoryPool struct {
	txQueue list.List
	txIndex map[string]*list.Element
	mux     sync.RWMutex
}

func NewMemoryPool() *MemoryPool {
	return &MemoryPool{
		txIndex: make(map[string]*list.Element),
	}
}

func (mp *MemoryPool) PushBackTx(tx *transactions.Transaction) {
	mp.mux.Lock()
	defer mp.mux.Unlock()
	queueNode := mp.txQueue.PushBack(tx)
	txHashString := hex.EncodeToString(tx.Hash())
	mp.txIndex[txHashString] = queueNode
}

func (mp *MemoryPool) PushFrontTx(tx *transactions.Transaction) {
	mp.mux.Lock()
	defer mp.mux.Unlock()
	queueNode := mp.txQueue.PushFront(tx)
	txHashString := hex.EncodeToString(tx.Hash())
	mp.txIndex[txHashString] = queueNode
}

func (mp *MemoryPool) DeleteTx(txHash []byte) {
	mp.mux.Lock()
	defer mp.mux.Unlock()
	txHashString := hex.EncodeToString(txHash)
	if queueNode, exists := mp.txIndex[txHashString]; exists {
		mp.txQueue.Remove(queueNode)
		delete(mp.txIndex, txHashString)
	}
}

func (mp *MemoryPool) GetTx(txHash []byte) *transactions.Transaction {
	mp.mux.RLock()
	defer mp.mux.RUnlock()
	txHashString := hex.EncodeToString(txHash)
	if queueNode, exists := mp.txIndex[txHashString]; exists {
		return queueNode.Value.(*transactions.Transaction)
	}
	return nil
}

func (mp *MemoryPool) FillBlockWithTxs(block *blockchain.Block) {
	mp.mux.RLock()
	defer mp.mux.RUnlock()
	for txQueueElement := mp.txQueue.Front(); txQueueElement != nil; txQueueElement = txQueueElement.Next() {
		tx := txQueueElement.Value.(*transactions.Transaction)
		addedTransaction := block.AddTransaction(tx)
		if !addedTransaction {
			break
		}
	}
}
