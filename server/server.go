package server

import (
	"encoding/hex"
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
	memoryPool      map[string]*transactions.Transaction
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
		memoryPool:   make(map[string]*transactions.Transaction),
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

	if server.blockInProgress == nil {
		server.memoryPool[hex.EncodeToString(tx.Hash())] = &tx
		return nil
	}

	addedTxToBlock := server.blockInProgress.AddTransaction(&tx)

	if !addedTxToBlock {
		server.memoryPool[hex.EncodeToString(tx.Hash())] = &tx
	}

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
			if server.blockInProgress != nil {
				if server.blockInProgress.Header.Height == server.bc.Height()+1 {
					server.mu.Lock()
					err := server.bc.AddBlock(server.blockInProgress)
					server.mu.Unlock()
					if err != nil {
						log.Panic("Error adding block to blockchain: ", err)
					}
					blockInProgressHash := server.blockInProgress.GetBlockHeaderHash()
					server.BroadcastObjects(INV, objectEntries{
						blockEntries: [][]byte{blockInProgressHash[:]},
					})
				} else {
					for _, tx := range server.blockInProgress.Transactions {
						server.memoryPool[hex.EncodeToString(tx.Hash())] = tx
					}
				}
			}

			server.blockInProgress = blockchain.NewBlock(
				[]*transactions.Transaction{transactions.NewCoinbaseTX(server.minerAddress)},
				server.bc.LastBlockHash(),
				server.bc.Height()+1,
			)

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
