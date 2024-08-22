package server

import (
	"github.com/pedrogomes29/blockchain_node/blockchain"
	"github.com/pedrogomes29/blockchain_node/transactions"
)

func (server *Server) AddBlockInProgressToBC() {
	server.mu.Lock()
	defer server.mu.Unlock()
	err := server.bc.AddBlock(server.blockInProgress)
	if err != nil {
		//TODO: error handling
	}
}

func (server *Server) POW() {

	newBlock := blockchain.NewBlock(
		[]*transactions.Transaction{transactions.NewCoinbaseTX(server.minerAddress)},
		server.bc.LastBlockHash(),
		server.bc.Height()+1,
	)

	server.memoryPool.FillBlockWithTxs(newBlock)

	server.blockInProgress = newBlock
	minedBlock := server.blockInProgress.POW(server.miningChan)

	if minedBlock {
		server.AddBlockInProgressToBC()
		blockInProgressHash := server.blockInProgress.GetBlockHeaderHash()
		server.BroadcastObjects(INV, objectEntries{
			blockEntries: [][]byte{blockInProgressHash[:]},
		})
	}

}
func (server *Server) POWLoop() {
	for {
		server.POW()
	}
}
