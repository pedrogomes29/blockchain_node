package server

import (
	"fmt"

	"github.com/pedrogomes29/blockchain_node/blockchain"
	"github.com/pedrogomes29/blockchain_node/transactions"
)

func (server *Server) AddBlockInProgToBC() error {
	server.mu.Lock()
	defer server.mu.Unlock()
	return server.AddBlockToBc(server.blockInProgress)
}

func (server *Server) POW() {

	newBlock := blockchain.NewBlock(
		[]*transactions.Transaction{transactions.NewCoinbaseTX(server.minerAddress)},
		server.bc.LastBlockHash(),
		server.bc.Height()+1,
	)

	newBlock.FillWithTxs(server.memoryPool)

	server.blockInProgress = newBlock
	minedBlock := server.blockInProgress.POW(server.miningChan)

	if minedBlock {
		err := server.AddBlockInProgToBC()

		if err != nil {
			fmt.Println("Error adding block in progress to blockchain")
			fmt.Println(err.Error())
		}
		blockInProgressHash := server.blockInProgress.GetBlockHeaderHash()
		server.BroadcastObjects(INV, objectEntries{
			blockEntries: [][]byte{blockInProgressHash},
		})
	}
}
func (server *Server) POWLoop() {
	for {
		server.POW()
	}
}
