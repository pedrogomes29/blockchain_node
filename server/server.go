package server

import (
	"sync"
	"time"

	"github.com/pedrogomes29/blockchain_node/blockchain"
	"github.com/pedrogomes29/blockchain_node/blockchain_errors"
	"github.com/pedrogomes29/blockchain_node/memory_pool"
	"github.com/pedrogomes29/blockchain_node/transactions"
)

type Server struct {
	bc              *blockchain.Blockchain
	minerAddress    string
	blockInProgress *blockchain.Block
	memoryPool      *memory_pool.MemoryPool	
	peers           map[string]*peer
	commands        chan command
	miningChan      chan struct{}
	mu              sync.Mutex
}

func NewServer(minerAddress string, seedAddrs []string) *Server {
	miningChan := make(chan struct{})
	server := &Server{
		bc:           blockchain.NewBlockchain(miningChan, minerAddress),
		minerAddress: minerAddress,
		memoryPool:   memory_pool.NewMemoryPool(),
		peers:        make(map[string]*peer),
		commands:     make(chan command),
		miningChan:   miningChan,
	}

	for _, seedAddres := range seedAddrs {
		server.ConnectToAddress(seedAddres)
	}
	return server
}

func (server *Server) AddTxToMemPool(tx transactions.Transaction) error {
	if !tx.VerifyInputSignatures(server.bc.ChainstateDB) {
		return &blockchain_errors.ErrInvalidTxInputSignature{}
	}

	server.memoryPool.PushBackTx(&tx)
	_ = server.blockInProgress.AddTransaction(&tx)

	return nil
}

func (server *Server) FindUTXOs(pubKeyHash []byte) ([]transactions.TXOutput, error) {
	return server.bc.FindUTXOs(pubKeyHash)
}

func (server *Server) FindSpendableUTXOs(pubKeyHash []byte, amount int) (int, map[string][]int, error) {
	return server.bc.FindSpendableUTXOs(pubKeyHash, amount)
}

func (server *Server) Run() {

	go server.HandleTcpCommands()
	go server.ListenForTcpConnections()
	go server.POWLoop()
	/*
		r := gin.Default()
		server.AddWalletRoutes(r)

		// Start the HTTP server
		if err := r.Run(":8080"); err != nil {
			panic("Failed to run server: " + err.Error())
		}
	*/

	for {
		time.Sleep(time.Minute)
	}
}
