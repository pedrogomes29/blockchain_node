package blockchain

import (
	"log"

	"github.com/pedrogomes29/blockchain/transactions"
)


func (bc *Blockchain) getBlocks() []*Block{
	blockBytes, err := bc.blocksDB.Get(bc.LastBlockHash,nil)
	if err != nil {
		log.Panic(err)
	}
	block := DeserializeBlock(blockBytes)
	blocks := []*Block{block}
	

	for (len(block.Header.PrevBlockHeaderHash) > 0){
		prevBlockHash := block.Header.PrevBlockHeaderHash
		blockBytes, err = bc.blocksDB.Get(prevBlockHash,nil)
		if err != nil {
			log.Panic(err)
		}
		block = DeserializeBlock(blockBytes)
		blocks = append([]*Block{block}, blocks...)
	}

	return blocks;
}

func (bc *Blockchain) ReindexUTXOs(){
	blocks := bc.getBlocks()

	for _, block := range blocks {
		for _, tx := range block.Transactions{
			txHash := tx.Hash()
			for _, txInput := range tx.Vin{
				//TODO: Check that txInput >= txOutput, check PubKey and Signature
				inputTxHash := txInput.Txid
				inputTxUTXObytes, err := bc.chainstateDB.Get(inputTxHash, nil)
				inputTxUTXOs := transactions.DeserializeUTXOs(inputTxUTXObytes)
				if err != nil {
					log.Panic(err)
				}

				_, ok := inputTxUTXOs[txInput.OutIndex]
				if !ok{
					log.Panic("Invalid transaction, spending from already used transaction output")
				}

				delete(inputTxUTXOs,txInput.OutIndex)
				err = bc.chainstateDB.Put(inputTxHash, inputTxUTXOs.Serialize(), nil)
				if err != nil {
					log.Panic(err)
				}
			}

			txUTXOs := make(transactions.UTXOs)
			for i, txoutput := range tx.Vout {
				txUTXOs[i] = txoutput
			}
			err := bc.chainstateDB.Put(txHash, txUTXOs.Serialize(), nil)
			if err != nil {
				log.Panic(err)
			}
		}
	}
}


func (bc * Blockchain) FindUTXOs(pubKeyHash []byte)  ([]transactions.TXOutput,error) {
	var UTXOs []transactions.TXOutput 
	txUTXOsIter := bc.chainstateDB.NewIterator(nil, nil)
	for txUTXOsIter.Next() {
		txUTXObytes := txUTXOsIter.Value()
		txUTXOs := transactions.DeserializeUTXOs(txUTXObytes)
		for _, UTXO := range txUTXOs{
			if UTXO.IsLockedWithKey(pubKeyHash){
				UTXOs = append(UTXOs, UTXO)
			}
		}
	}
	txUTXOsIter.Release()
	err := txUTXOsIter.Error()
	return UTXOs,err
}