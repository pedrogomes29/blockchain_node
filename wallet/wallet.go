package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"

	"github.com/btcsuite/btcd/btcutil/base58"
	"github.com/pedrogomes29/blockchain/server"
	"github.com/pedrogomes29/blockchain/transactions"
	"golang.org/x/crypto/ripemd160"
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
				PubKey:wallet.PublicKey(),
			}
			//TODO: sign transaction
			inputs = append(inputs, input)
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
	publicSHA256 := sha256.Sum256(wallet.PublicKey())

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}
	publicRIPEMD160 := RIPEMD160Hasher.Sum(nil)

	return publicRIPEMD160
}

func (wallet *Wallet) Address() string{
	pubKeyHash := wallet.PublicKeyHash()
	return base58.CheckEncode(pubKeyHash,0x00)
}