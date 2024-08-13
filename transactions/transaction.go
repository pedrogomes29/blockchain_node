package transactions

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
)

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

const subsidy = 10 //TODO calculate dynamically given number of blocks (deflationary)


func NewCoinbaseTX(receiverAddress string) *Transaction {
	txin := TXInput{[]byte{}, -1, nil, []byte{}}
	txout, _ := NewTXOutput(subsidy, receiverAddress) //TODO: Handle invalid bitcoin address error
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}}
	tx.ID = tx.Hash()
	return &tx
}


func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}


func (tx Transaction) Hash() []byte {
	var hash [32]byte

	hash = sha256.Sum256(tx.Serialize())

	return hash[:]
}
