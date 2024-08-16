package server

import (
	"fmt"
	"sync"

	"github.com/pedrogomes29/blockchain/blockchain"
	"github.com/pedrogomes29/blockchain/transactions"
)


type Server struct{
	Bc *blockchain.Blockchain
	minerAddress string
	blockInProgress *blockchain.Block
	mu               sync.Mutex
}

func NewServer(minerAddress string) *Server{
	return &Server{
		Bc: blockchain.NewBlockchain(minerAddress),
		minerAddress: minerAddress,
	}
}

func (server *Server) AddTransaction(tx *transactions.Transaction){
	if(!tx.VerifyInputSignatures(server.Bc.ChainstateDB)){
		fmt.Println("Transaction inputs have at least one invalid signature")
		return;
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	if server.blockInProgress != nil{
		server.blockInProgress.Transactions = append(server.blockInProgress.Transactions, tx)
	}
}


func (server *Server) Run(){
	done := make(chan struct{})
	for {
		server.mu.Lock()
		if server.blockInProgress!=nil {
			server.Bc.AddBlock(server.blockInProgress)

		}
		server.blockInProgress = blockchain.NewBlock(
			[]*transactions.Transaction{transactions.NewCoinbaseTX(server.minerAddress)},
			server.Bc.LastBlockHash,
		)
		server.mu.Unlock()

		go server.blockInProgress.POW(done)
		<-done
	}
}