package transactions

import (
	"bytes"
	"encoding/gob"
	"log"
)

type UTXOs map[int]TXOutput

func (utxos UTXOs) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(utxos)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func DeserializeUTXOs(utxoBytes []byte) UTXOs {
	var utxos UTXOs

	decoder := gob.NewDecoder(bytes.NewReader(utxoBytes))
	err := decoder.Decode(&utxos)
	if err != nil {
		log.Panic(err)
	}

	return utxos
}