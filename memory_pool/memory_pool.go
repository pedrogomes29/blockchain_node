package memory_pool

import (
	"container/list"
	"encoding/hex"

	"github.com/pedrogomes29/blockchain_node/transactions"
)

type MemoryPool struct {
	txQueue list.List
	txIndex map[string]*list.Element
}

func NewMemoryPool() *MemoryPool{
	return &MemoryPool{
		txIndex: make(map[string]*list.Element),
	}
}

func (mp *MemoryPool) PushBackTx(tx *transactions.Transaction){
	queueNode := mp.txQueue.PushBack(tx)
	txHashString := hex.EncodeToString(tx.Hash())
	mp.txIndex[txHashString] = queueNode
}

func (mp *MemoryPool) PushFrontTx(tx *transactions.Transaction){
	queueNode := mp.txQueue.PushFront(tx)
	txHashString := hex.EncodeToString(tx.Hash())
	mp.txIndex[txHashString] = queueNode
}

func (mp *MemoryPool) DeleteTx(txHash []byte){
	txHashString := hex.EncodeToString(txHash)
	if queueNode, exists := mp.txIndex[txHashString]; exists {
		mp.txQueue.Remove(queueNode)
		delete(mp.txIndex, txHashString)
	}
}

func (mp *MemoryPool) GetTx(txHash []byte) *transactions.Transaction{
	txHashString := hex.EncodeToString(txHash)
	if queueNode, exists := mp.txIndex[txHashString]; exists {
		return queueNode.Value.(*transactions.Transaction)
	}
	return nil
}




