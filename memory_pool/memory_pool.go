package memory_pool

import (
	"container/list"
	"encoding/hex"
	"sync"

	"github.com/pedrogomes29/blockchain_node/blockchain_errors"
	"github.com/pedrogomes29/blockchain_node/transactions"
)

type inputUTXOs map[int]string //maps utxo to transaction spending that utxo

type MemoryPool struct {
	txQueue    *list.List
	txIndex    map[string]*list.Element
	spentUTXOs map[string]inputUTXOs //maps transaction id to that transaction's UTXOs being spent
	mux        *sync.RWMutex
}

func NewMemoryPool() *MemoryPool {
	return &MemoryPool{
		txQueue:    list.New(),
		txIndex:    make(map[string]*list.Element),
		spentUTXOs: make(map[string]inputUTXOs),
		mux:        &sync.RWMutex{},
	}
}

func (mp *MemoryPool) GetRWMutex() *sync.RWMutex {
	return mp.mux
}

func (mp *MemoryPool) GetTxQueue() *list.List {
	return mp.txQueue
}

func (mp *MemoryPool) isAnyInputAlreadySpent(tx *transactions.Transaction) bool {
	for _, txInput := range tx.Vin {
		inputTxHash := hex.EncodeToString(txInput.Txid)
		if utxoInputs, exists := mp.spentUTXOs[inputTxHash]; exists {
			if _, spent := utxoInputs[txInput.OutIndex]; spent {
				return true
			}
		}
	}
	return false
}

func (mp *MemoryPool) markInputsAsSpent(tx *transactions.Transaction) {
	for _, txInput := range tx.Vin {
		inputTxHash := hex.EncodeToString(txInput.Txid)
		if _, exists := mp.spentUTXOs[inputTxHash]; !exists {
			mp.spentUTXOs[inputTxHash] = make(inputUTXOs)
		}

		mp.spentUTXOs[inputTxHash][txInput.OutIndex] = hex.EncodeToString(tx.Hash())
	}
}

func (mp *MemoryPool) unmarkUTXOInputsAsSpent(tx *transactions.Transaction) {
	for _, txInput := range tx.Vin {
		inputTxHash := hex.EncodeToString(txInput.Txid)
		delete(mp.spentUTXOs, inputTxHash)
	}
}

func (mp *MemoryPool) deleteTxsSpendingSameUTXOs(tx *transactions.Transaction) error {
	for _, txInput := range tx.Vin { //deletes all transactions which spend at least one UTXO in common
		inputTxHash := hex.EncodeToString(txInput.Txid)
		if utxoInputs, exists := mp.spentUTXOs[inputTxHash]; exists {
			if txWhichSpent, spent := utxoInputs[txInput.OutIndex]; spent {
				txWhichSpentHash, err := hex.DecodeString(txWhichSpent)
				if err != nil {
					return err
				}
				mp.deleteTx(txWhichSpentHash)
			}
		}

	}
	return nil
}

func (mp *MemoryPool) DeleteTxsSpendingFromTxUTXOsWithLock(tx *transactions.Transaction) error {
	mp.mux.Lock()
	defer mp.mux.Unlock()

	txHash := hex.EncodeToString(tx.Hash())

	if utxoInputs, exists := mp.spentUTXOs[txHash]; exists {
		for _, inputTxHashEncoded := range utxoInputs {
			inputTxHash, err := hex.DecodeString(inputTxHashEncoded)
			if err != nil {
				return err
			}
			mp.deleteTx(inputTxHash)
		}
	}
	return nil
}

func (mp *MemoryPool) pushBackTx(tx *transactions.Transaction) error {
	if tx := mp.getTx(tx.Hash()); tx != nil {
		return nil //tx already in queue, do nothing to maintain queue spot (and not be pushed to the back)
	}

	if mp.isAnyInputAlreadySpent(tx) {
		return &blockchain_errors.ErrInvalidInputUTXO{}
	}

	queueNode := mp.txQueue.PushBack(tx)
	txHashString := hex.EncodeToString(tx.Hash())
	mp.txIndex[txHashString] = queueNode
	mp.markInputsAsSpent(tx)

	return nil
}

func (mp *MemoryPool) PushBackTxWithLock(tx *transactions.Transaction) error {
	mp.mux.Lock()
	defer mp.mux.Unlock()

	return mp.pushBackTx(tx)
}

func (mp *MemoryPool) pushFrontTx(tx *transactions.Transaction) error {
	if tx := mp.getTx(tx.Hash()); tx != nil {
		mp.deleteTx(tx.Hash()) //tx already in queue, delete current position to push to the front of the queue
	}

	err := mp.deleteTxsSpendingSameUTXOs(tx)
	if err != nil {
		return err
	}

	queueNode := mp.txQueue.PushFront(tx)
	txHashString := hex.EncodeToString(tx.Hash())
	mp.txIndex[txHashString] = queueNode
	mp.markInputsAsSpent(tx)

	return nil
}

func (mp *MemoryPool) PushFrontTxWithLock(tx *transactions.Transaction) error {
	mp.mux.Lock()
	defer mp.mux.Unlock()
	return mp.pushFrontTx(tx)

}

func (mp *MemoryPool) DeleteTxWithLock(txHash []byte) {
	mp.mux.Lock()
	defer mp.mux.Unlock()
	mp.deleteTx(txHash)
}

func (mp *MemoryPool) deleteTx(txHash []byte) {
	txHashString := hex.EncodeToString(txHash)
	if queueNode, exists := mp.txIndex[txHashString]; exists {
		tx := queueNode.Value.(*transactions.Transaction)
		mp.txQueue.Remove(queueNode)
		delete(mp.txIndex, txHashString)
		mp.unmarkUTXOInputsAsSpent(tx)
	}
}

func (mp *MemoryPool) getTx(txHash []byte) *transactions.Transaction {
	txHashString := hex.EncodeToString(txHash)
	if queueNode, exists := mp.txIndex[txHashString]; exists {
		return queueNode.Value.(*transactions.Transaction)
	}
	return nil
}

func (mp *MemoryPool) GetTxWithLock(txHash []byte) *transactions.Transaction {
	mp.mux.RLock()
	defer mp.mux.RUnlock()
	return mp.getTx(txHash)
}
