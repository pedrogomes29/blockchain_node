package server

import (
	"log"
	"sync"
	"time"

	"github.com/pedrogomes29/blockchain_node/blockchain"
	"github.com/pedrogomes29/blockchain_node/blockchain_errors"
	"github.com/pedrogomes29/blockchain_node/transactions"
)

type Server struct {
	bc              *blockchain.Blockchain
	minerAddress    string
	blockInProgress *blockchain.Block
	peers           map[string]*peer
	commands        chan command
	miningChan      chan int
	mu              sync.Mutex
}

func NewServer(minerAddress string, seedAddrs []string) *Server {
	miningChan := make(chan int)
	server := &Server{
		bc:           blockchain.NewBlockchain(miningChan, minerAddress),
		minerAddress: minerAddress,
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

	server.mu.Lock()
	defer server.mu.Unlock()

	if server.blockInProgress != nil {
		server.blockInProgress.Transactions = append(server.blockInProgress.Transactions, &tx)
	}

	server.blockInProgress.Header.MerkleRootHash = server.blockInProgress.MerkleRootHash()

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

	go func() {
		for {
			server.mu.Lock()
			if server.blockInProgress != nil && server.blockInProgress.Header.Height == server.bc.Height+1 {
				err := server.bc.AddBlock(server.blockInProgress)
				if err != nil {
					log.Panic("Error adding block to blockchain: ", err)
				}
				blockInProgressHash := server.blockInProgress.GetBlockHeaderHash()
				server.BroadcastObjects(INV, []objectEntry{
					{BLOCK, blockInProgressHash[:]},
				})
			}

			server.blockInProgress = blockchain.NewBlock(
				[]*transactions.Transaction{transactions.NewCoinbaseTX(server.minerAddress)},
				server.bc.LastBlockHash,
				server.bc.Height+1,
			)
			server.mu.Unlock()

			server.blockInProgress.POW(server.miningChan)
		}
	}()

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
