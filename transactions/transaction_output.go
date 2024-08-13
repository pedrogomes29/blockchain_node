package transactions

import (
	"fmt"

	"github.com/btcsuite/btcd/btcutil/base58"

	"errors"
)

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

var ErrInvalidAddress = errors.New("Invalid Bitcoin Address")


func NewTXOutput(value int, address string) (txOutput *TXOutput, err error) {
	pubKeyHash, version , err := base58.CheckDecode(address)
	if err != nil || version != 0x00{
		fmt.Println("Invalid bitcoin address")
		return nil, ErrInvalidAddress
	}
	txo := &TXOutput{value, pubKeyHash}
	return txo, nil
}