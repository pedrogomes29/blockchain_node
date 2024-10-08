package blockchain

import (
	"encoding/hex"

	"github.com/pedrogomes29/blockchain_node/transactions"
	"github.com/syndtr/goleveldb/leveldb/util"
)

func (bc *Blockchain) ReindexUTXOs() {
	blocks := bc.GetBlocksStartingAtHash([]byte{})

	for _, block := range blocks {
		for _, tx := range block.Transactions {
			tx.IndexUTXOs(bc.ChainstateDB)
		}
	}
}

func (bc *Blockchain) FindUTXOs(pubKeyHash []byte) ([]transactions.TXOutput, error) {
	var UTXOs []transactions.TXOutput
	txUTXOsIter := bc.ChainstateDB.NewIterator(util.BytesPrefix([]byte(transactions.UTXO_PREFIX)), nil)
	for txUTXOsIter.Next() {
		txUTXObytes := txUTXOsIter.Value()
		txUTXOs := transactions.DeserializeUTXOs(txUTXObytes)
		for _, UTXO := range txUTXOs {
			if UTXO.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, UTXO)
			}
		}
	}
	txUTXOsIter.Release()
	err := txUTXOsIter.Error()
	return UTXOs, err
}

func (bc *Blockchain) FindSpendableUTXOs(pubKeyHash []byte, amount int) (int, map[string][]int, error) {
	UTXOs := make(map[string][]int)
	utxoTotalAmount := 0
	txUTXOsIter := bc.ChainstateDB.NewIterator(util.BytesPrefix([]byte(transactions.UTXO_PREFIX)), nil)
	for txUTXOsIter.Next() {
		txHash := txUTXOsIter.Key()[len(transactions.UTXO_PREFIX):]
		txUTXObytes := txUTXOsIter.Value()
		txUTXOs := transactions.DeserializeUTXOs(txUTXObytes)
		for utxoIndex, UTXO := range txUTXOs {
			if UTXO.IsLockedWithKey(pubKeyHash) {
				utxoTotalAmount += UTXO.Value
				UTXOs[hex.EncodeToString(txHash)] = append(UTXOs[hex.EncodeToString(txHash)], utxoIndex)
			}
		}
	}
	txUTXOsIter.Release()
	err := txUTXOsIter.Error()
	return utxoTotalAmount, UTXOs, err
}
