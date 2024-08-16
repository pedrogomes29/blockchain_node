package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"log"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/pedrogomes29/blockchain/server"
	"github.com/pedrogomes29/blockchain/transactions"
	"github.com/pedrogomes29/blockchain/utils"
)

type Wallet struct{
	server *server.Server
	privateKey ecdsa.PrivateKey
}

func NewWalletAndPrivateKey() *Wallet {
	privateKey := newPrivateKey()
	wallet := Wallet{
		privateKey: privateKey,
	}
	return &wallet
}

func NewWallet(privateKey ecdsa.PrivateKey) *Wallet {
	wallet := Wallet{
		privateKey: privateKey,
	}
	return &wallet
}

func newPrivateKey() ecdsa.PrivateKey{
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}
	return *private
}

func (wallet *Wallet) PublicKey() []byte{
	privateKey := wallet.privateKey
	publicKey := privateKey.PublicKey

	return append(publicKey.X.Bytes(), publicKey.Y.Bytes()...)
}


func (wallet *Wallet) SetServer(server *server.Server){
	wallet.server = server
}


func (wallet *Wallet) generateTxToAddress(toAddress string, amount int) *transactions.Transaction{
	var inputs []transactions.TXInput
	var outputs []transactions.TXOutput
	inputIdxsToSign := make(map[int]struct{})

	publicKeyHash := wallet.PublicKeyHash()
	utxosTotal, spendableUTXOs, err := wallet.server.Bc.FindSpendableUTXOs(publicKeyHash,amount)
	if err!=nil{
		log.Panic(err)
	}
	if utxosTotal < amount {
		log.Panic("ERROR: Not enough funds")
	}

	for txHashString, outs := range spendableUTXOs {
		txHash, err := hex.DecodeString(txHashString)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := transactions.TXInput{
				Txid: txHash,
				OutIndex: out,
				Signature: nil,
				PubKey: wallet.PublicKey(),
			}
			inputs = append(inputs, input)
			inputIdxsToSign[len(inputs)-1] = struct{}{}
		}
	}
	transactionToAddress, err := transactions.NewTXOutput(amount,toAddress)
	if err != nil {
		log.Panic(err)
	}
	outputs = append(outputs, *transactionToAddress)

	if(utxosTotal > amount){
		transactionChange, err := transactions.NewTXOutput(utxosTotal - amount, wallet.Address())
		if err != nil {
			log.Panic(err)
		}
		outputs = append(outputs, *transactionChange)
	}

	transaction := &transactions.Transaction{
		ID: nil, 
		Vin: inputs,
		Vout: outputs,
		IsCoinbase: false,
	}
	transaction.ID = transaction.Hash()

	wallet.SignTransactionInputs(transaction,inputIdxsToSign)

	return transaction
}

func (wallet *Wallet) SendToAddress(toAddress string, amount int){
	tx := wallet.generateTxToAddress(toAddress, amount)
	wallet.server.AddTransaction(tx)
}

func (wallet *Wallet) GetBalance() int{
	UTXOs, _ := wallet.server.Bc.FindUTXOs(wallet.PublicKeyHash())
	balance := 0
	for _, out := range UTXOs {
		balance += out.Value
	}
	return balance
}

func (wallet *Wallet) PublicKeyHash()[]byte{
	return utils.HashPublicKey(wallet.PublicKey())
}

func (wallet *Wallet) Address() string{
	pubKeyHash := wallet.PublicKeyHash()
	return base58.CheckEncode(pubKeyHash,0x00)
}

func (wallet *Wallet) SignTransactionInputs(tx *transactions.Transaction, inputIdxsToSign map[int]struct{}){
	txCopy := tx.TrimmedCopy()
	for inputIdxToSign := range inputIdxsToSign{
		r, s, err := ecdsa.Sign(rand.Reader, &wallet.privateKey, txCopy.Hash())
		if err!=nil{
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inputIdxToSign].Signature = signature
	}
}