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

	"github.com/pedrogomes29/blockchain_node/utils"
	"github.com/syndtr/goleveldb/leveldb"
)

type Transaction struct {
	Vin        []TXInput
	Vout       []TXOutput
	IsCoinbase bool
}


const UTXO_PREFIX string = "utxo:"
const REV_UTXO_PREFIX string = "rev:"

const subsidy = 10 //TODO calculate dynamically given number of blocks (deflationary)

func NewCoinbaseTX(receiverAddress string) *Transaction {
	txout, err := NewTXOutput(subsidy, receiverAddress)
	if err != nil {
		log.Panic(err)
	}
	txin := TXInput{[]byte{}, -1, nil, []byte(utils.GenerateRandomString(20))}
	tx := Transaction{[]TXInput{txin}, []TXOutput{*txout}, true}
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

func Deserialize(data []byte) *Transaction {
	var tx Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&tx)
	if err != nil {
		log.Panic(err)
	}

	return &tx
}

func (tx Transaction) Hash() []byte {
	hash := sha256.Sum256(tx.Serialize())

	return hash[:]
}

func (tx Transaction) IndexUTXOs(chainstateDB *leveldb.DB) error {
	if tx.IsCoinbase {
		txUTXOs := make(UTXOs)
		for i, txoutput := range tx.Vout {
			txUTXOs[i] = txoutput
		}
		err := chainstateDB.Put(append([]byte(UTXO_PREFIX),tx.Hash()...), txUTXOs.Serialize(), nil)
		if err != nil {
			return err
		}
		return nil
	}

	if !tx.VerifyInputSignatures(chainstateDB) {
		return errors.New("Transaction inputs have at least one invalid signature")
	}

	txInputTotal := 0

	//stores updated UTXOs temporarily to only update after verifying transaction is valid
	inputTxsUpdatedUTXOs := make(map[string]UTXOs)

	//stores UTXOs that were deleted from the index temporarily to only update after verifying transaction is valid
	inputTxsSpentUTXOs := make(map[string]UTXOs)

	for _, txInput := range tx.Vin {
		inputTxHash := txInput.Txid
		inputTxUTXObytes, err := chainstateDB.Get(append([]byte(UTXO_PREFIX), inputTxHash...), nil)
		if err != nil {
			return err
		}
		inputTxUTXOs := DeserializeUTXOs(inputTxUTXObytes)
		prevUTXO, isUTXO := inputTxUTXOs[txInput.OutIndex]
		if !isUTXO { //if spent UTXO
			return errors.New("invalid transaction, spending from already used transaction output")
		}

		txInputTotal += prevUTXO.Value

		inputHashString := hex.EncodeToString(inputTxHash)
		_, spentUTXOFromInputTx := inputTxsSpentUTXOs[inputHashString]
		if !spentUTXOFromInputTx {
			inputTxsSpentUTXOs[inputHashString] = make(UTXOs)
		}
		inputTxsSpentUTXOs[inputHashString][txInput.OutIndex] = inputTxUTXOs[txInput.OutIndex]

		delete(inputTxUTXOs, txInput.OutIndex)
		inputTxsUpdatedUTXOs[inputHashString] = inputTxUTXOs
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
	inputTxsUpdatedUTXOs[hex.EncodeToString(tx.Hash())] = txUTXOs

	for inputTxHashString := range inputTxsUpdatedUTXOs {
		inputTxHash, err := hex.DecodeString(inputTxHashString)
		inputTxUpdatedUTXOs := inputTxsUpdatedUTXOs[inputTxHashString]
		inputTxSpentUTXOs, spentUTXOFromInputTx := inputTxsSpentUTXOs[inputTxHashString]

		if err != nil {
			return err
		}
		err = chainstateDB.Put(append([]byte(UTXO_PREFIX), inputTxHash...), inputTxUpdatedUTXOs.Serialize(), nil) //store updated utxos
		if err != nil {
			return err
		}

		if spentUTXOFromInputTx {
			inputTxRevUTXOsKey := bytes.Join([][]byte{
				[]byte(REV_UTXO_PREFIX),
				tx.Hash(),
				[]byte(":"),
				inputTxHash,
			}, nil)

			//store the UTXOs that were spent to allow for reversing the transaction index
			err = chainstateDB.Put(inputTxRevUTXOsKey, inputTxSpentUTXOs.Serialize(), nil)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (tx Transaction) RevertUTXOIndex(chainstateDB *leveldb.DB) error {
	err := chainstateDB.Delete(append([]byte(UTXO_PREFIX), tx.Hash()...), nil) //deletes UTXOs of the current transaction
	if err != nil {
		return err
	}

	for _, txInput := range tx.Vin {
		inputTxHash := txInput.Txid
		inputTxUTXObytes, err := chainstateDB.Get(append([]byte(UTXO_PREFIX), inputTxHash...), nil)
		if err != nil {
			return err
		}
		inputTxUTXOs := DeserializeUTXOs(inputTxUTXObytes)

		inputTxRevUTXOsKey := bytes.Join([][]byte{
			[]byte(REV_UTXO_PREFIX),
			tx.Hash(),
			[]byte(":"),
			inputTxHash,
		}, nil)

		inputTxRevUTXOsbytes, err := chainstateDB.Get(inputTxRevUTXOsKey, nil)
		if err == leveldb.ErrNotFound { //Nothing to reverse for this transaction input
			continue
		}
		if err != nil {
			return err
		}

		//UTXOs that were spent in the current transaction (reverse index)
		inputTxRevUTXOs := DeserializeUTXOs(inputTxRevUTXOsbytes)

		for outIdx, utxo := range inputTxRevUTXOs {
			inputTxUTXOs[outIdx] = utxo //add utxos back
		}

		err = chainstateDB.Put(append([]byte(UTXO_PREFIX), inputTxHash...), inputTxUTXOs.Serialize(), nil) //store updated UTXOs
		if err != nil {
			return err
		}

		err = chainstateDB.Delete(inputTxRevUTXOsKey, nil) //delete reverse index for this input transaction
		if err != nil {
			return err
		}
	}

	return nil
}

func (tx Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, txIn := range tx.Vin {
		inputs = append(inputs, TXInput{
			txIn.Txid,
			txIn.OutIndex,
			nil,
			nil,
		})
	}
	for _, txOut := range tx.Vout {
		outputs = append(outputs, TXOutput{txOut.Value, txOut.PubKeyHash})
	}

	txTrimmed := Transaction{
		inputs,
		outputs,
		tx.IsCoinbase,
	}

	return txTrimmed
}

func (tx Transaction) VerifyInputSignatures(chainstateDB *leveldb.DB) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for _, txIn := range tx.Vin {
		inputTxUTXObytes, err := chainstateDB.Get(append([]byte(UTXO_PREFIX), txIn.Txid...), nil)
		if err != nil {
			return false
		}
		inputTxUTXOs := DeserializeUTXOs(inputTxUTXObytes)
		inputTxUTXO := inputTxUTXOs[txIn.OutIndex]

		if !bytes.Equal(utils.HashPublicKey(txIn.PubKey), inputTxUTXO.PubKeyHash) {
			return false
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
