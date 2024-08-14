package transactions

import (
	"bytes"
	"fmt"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/pedrogomes29/blockchain/blockchain_errors"
)

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}


func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func NewTXOutput(value int, address string) (txOutput *TXOutput, err error) {
	pubKeyHash, version , err := base58.CheckDecode(address)
	if err != nil || version != 0x00{
		fmt.Println("Invalid bitcoin address")
		return nil, &blockchain_errors.ErrInvalidAddress{}
	}
	txo := &TXOutput{value, pubKeyHash}
	return txo, nil
}