package server

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/pedrogomes29/blockchain/blockchain"
	"github.com/pedrogomes29/blockchain/transactions"
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

func (server *Server) AddTransactionHandler(c *gin.Context) {
	var tx transactions.Transaction

	if err := c.ShouldBindJSON(&tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction format"})
		return
	}

	if !tx.VerifyInputSignatures(server.bc.ChainstateDB) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction inputs have at least one invalid signature"})
		return
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	if server.blockInProgress != nil {
		server.blockInProgress.Transactions = append(server.blockInProgress.Transactions, &tx)
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Transaction added to mempool successfully"})
}

func (server *Server) FindUTXOsHandler(c *gin.Context) {
	var pubKeyHash []byte

	if err := c.ShouldBindJSON(&pubKeyHash); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid public key hash format"})
		return
	}

	utxos, err := server.bc.FindUTXOs(pubKeyHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding UTXOs"})
		return
	}

	c.JSON(http.StatusOK, utxos)
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
	r.POST("/transaction", server.AddTransactionHandler)
	r.GET("/utxos", server.FindUTXOsHandler)

	// Start the HTTP server
	if err := r.Run(":8080"); err != nil {
		panic("Failed to run server: " + err.Error())
	}
}
