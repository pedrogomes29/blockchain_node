package transactions

import (
	"bytes"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/pedrogomes29/blockchain_node/blockchain_errors"
)

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}

func NewTXOutput(value int, address string) (txOutput *TXOutput, err error) {
	pubKeyHash, version, err := base58.CheckDecode(address)
	if err != nil {
		return nil, err
	}
	if version != 0x00 {
		return nil, &blockchain_errors.ErrInvalidAddress{}
	}
	txo := &TXOutput{value, pubKeyHash}
	return txo, nil
}
