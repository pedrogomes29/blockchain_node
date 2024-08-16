package transactions

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"errors"
	"log"
	"math/big"

	"github.com/pedrogomes29/blockchain/utils"
	"github.com/syndtr/goleveldb/leveldb"
)

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
	IsCoinbase bool
}

const subsidy = 10 //TODO calculate dynamically given number of blocks (deflationary)


func NewCoinbaseTX(receiverAddress string) *Transaction {
	txout, err := NewTXOutput(subsidy, receiverAddress) //TODO: Handle invalid bitcoin address error
	if err!=nil{
		log.Panic(err)
	}
	txin := TXInput{[]byte{}, -1, nil, []byte(utils.GenerateRandomString(20))}
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{*txout}, true}
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
	hash := sha256.Sum256(tx.Serialize())

	return hash[:]
}

func (tx Transaction) IndexUTXOs(chainstateDB *leveldb.DB) error{
	if tx.IsCoinbase{
		txUTXOs := make(UTXOs)
		for i, txoutput := range tx.Vout {
			txUTXOs[i] = txoutput
		}
		err := chainstateDB.Put(tx.ID, txUTXOs.Serialize(), nil)
		if err != nil {
			return err;
		}
		return nil;
	}

	if !tx.VerifyInputSignatures(chainstateDB){
		return errors.New("Transaction inputs have at least one invalid signature")
	}

	txInputTotal := 0

	updatedUTXOs := make(map[string]UTXOs) //stores updated UTXOs temporarily to only update after verifying transaction is valid

	for _, txInput := range tx.Vin{
		inputTxHash := txInput.Txid
		inputTxUTXObytes, err := chainstateDB.Get(inputTxHash, nil)
		if err != nil {
			return err
		}
		inputTxUTXOs := DeserializeUTXOs(inputTxUTXObytes)
		//TODO: Verify user can 
		prevUTXO, ok := inputTxUTXOs[txInput.OutIndex]
		if !ok{
			return errors.New("invalid transaction, spending from already used transaction output")
		}

		txInputTotal += prevUTXO.Value


		delete(inputTxUTXOs,txInput.OutIndex)
		updatedUTXOs[hex.EncodeToString(inputTxHash)] = inputTxUTXOs
	}
	
	txUTXOs := make(UTXOs)
	txOutputTotal := 0
	for i, txoutput := range tx.Vout {
		txUTXOs[i] = txoutput
		txOutputTotal += txoutput.Value
	}

	if txInputTotal < txOutputTotal {
		return errors.New("invalid transaction, total output value is larger than total input value")
	}
	updatedUTXOs[hex.EncodeToString(tx.ID)] = txUTXOs

	for updatedTXHashString, utxos := range updatedUTXOs{
		updatedTXHash, err := hex.DecodeString(updatedTXHashString)
		if err != nil {
			return err;
		}
		err = chainstateDB.Put(updatedTXHash, utxos.Serialize(), nil)
		if err != nil {
			return err;
		}
	}

	return nil;
}

func (tx Transaction) TrimmedCopy() Transaction{
	var inputs []TXInput
	var outputs []TXOutput

	for _,txIn := range tx.Vin{
		inputs = append(inputs, TXInput{
			txIn.Txid,
			txIn.OutIndex,
			nil,
			nil,
		})
	}
	for _,txOut := range tx.Vout{
		outputs = append(outputs, TXOutput{ txOut.Value, txOut.PubKeyHash})
	}

	txTrimmed := Transaction{
		tx.ID,
		inputs,
		outputs,
		tx.IsCoinbase,
	}

	return txTrimmed
}

func (tx Transaction) VerifyInputSignatures(chainstateDB *leveldb.DB) bool{
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for _,txIn:= range tx.Vin{
		inputTxUTXObytes, err := chainstateDB.Get(txIn.Txid, nil)
		if err != nil {
			return false;
		}
		inputTxUTXOs := DeserializeUTXOs(inputTxUTXObytes)
		inputTxUTXO := inputTxUTXOs[txIn.OutIndex]

		if !bytes.Equal(utils.HashPublicKey(txIn.PubKey), inputTxUTXO.PubKeyHash){
			return false;
		}

		r := big.Int{}
		s := big.Int{}
		sigLen := len(txIn.Signature)
		r.SetBytes(txIn.Signature[:(sigLen / 2)])
		s.SetBytes(txIn.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(txIn.PubKey)
		x.SetBytes(txIn.PubKey[:(keyLen / 2)])
		y.SetBytes(txIn.PubKey[(keyLen / 2):])


		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, txCopy.Hash(), &r, &s) {
			return false
		}
	}
	return true
}