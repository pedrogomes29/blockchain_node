package server

import (
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/pedrogomes29/blockchain_node/blockchain"
	"github.com/pedrogomes29/blockchain_node/blockchain_errors"
	"github.com/pedrogomes29/blockchain_node/transactions"
)

type Server struct {
	bc              *blockchain.Blockchain
	minerAddress    string
	blockInProgress *blockchain.Block
	mu              sync.Mutex
}

func NewServer(minerAddress string) *Server {
	return &Server{
		bc:           blockchain.NewBlockchain(minerAddress),
		minerAddress: minerAddress,
	}
}

func (server *Server) AddTxToMemPool(tx transactions.Transaction) error{
	if !tx.VerifyInputSignatures(server.bc.ChainstateDB) {
		return &blockchain_errors.ErrInvalidTxInputSignature{}
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	if server.blockInProgress != nil {
		server.blockInProgress.Transactions = append(server.blockInProgress.Transactions, &tx)
	}

	return nil
}

func (server *Server) FindUTXOs(pubKeyHash []byte) ([]transactions.TXOutput, error){
	return server.bc.FindUTXOs(pubKeyHash)
}

func (server *Server) FindSpendableUTXOs(pubKeyHash []byte, amount int) (int, map[string][]int, error){
	return server.bc.FindSpendableUTXOs(pubKeyHash, amount)
}


func (server *Server) Run() {
	done := make(chan struct{})

	go func() {
		for {
			server.mu.Lock()
			if server.blockInProgress != nil {
				server.bc.AddBlock(server.blockInProgress)
			}
			server.blockInProgress = blockchain.NewBlock(
				[]*transactions.Transaction{transactions.NewCoinbaseTX(server.minerAddress)},
				server.bc.LastBlockHash,
			)
			server.mu.Unlock()

			go server.blockInProgress.POW(done)
			<-done
		}
	}()

	r := gin.Default()
	server.AddWalletRoutes(r)

	// Start the HTTP server
	if err := r.Run(":8080"); err != nil {
		panic("Failed to run server: " + err.Error())
	}
}
