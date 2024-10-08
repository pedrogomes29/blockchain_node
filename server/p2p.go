package server

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/pedrogomes29/blockchain_node/blockchain"
	"github.com/pedrogomes29/blockchain_node/blockchain_errors"
	"github.com/pedrogomes29/blockchain_node/transactions"
)

const BLOCKCHAIN_PORT string = "8333"
const BLOCK_CONFIRMATIONS int = 6

func (server *Server) ConnectToAddress(address string) {
	if _, ok := server.peers[address]; ok { //if address is already known
		return
	}

	conn, err := net.Dial("tcp", address+":"+BLOCKCHAIN_PORT)
	if err != nil {
		log.Panic("Error establishing connection: ", err)
		return
	}

	newPeer := &peer{
		conn:     conn,
		commands: server.commands,
	}

	//log.Printf("successfully established connection to a new peer: %s\n", address)

	newPeer.sendString("VERSION" + " " + strconv.Itoa(server.bc.Height()))

	go newPeer.ReadInput()
}

func (server *Server) ReceiveAddresses(addresses addrPayload) {
	for _, address := range addresses {
		server.ConnectToAddress(address)
	}
}

func (server *Server) SendGetBlocks(requestPeer *peer) {
	lastBlockHashes := server.bc.GetLastBlockHashes(BLOCK_CONFIRMATIONS)
	var sb strings.Builder
	sb.WriteString("GET_BLOCKS")
	for _, blockHash := range lastBlockHashes {
		sb.WriteString(" " + hex.EncodeToString(blockHash))
	}
	requestPeer.sendString(sb.String())
}

func (server *Server) ReceiveGetBlocks(requestPeer *peer, payload getBlocksPayload) {
	var block *blockchain.Block
	var highestCommonBlockHash []byte
	for _, highestCommonBlockHash = range payload {
		block = server.bc.GetBlock(highestCommonBlockHash)
		if block != nil {
			break
		}
	}

	if block == nil {
		highestCommonBlockHash = []byte{}
	}

	var unsharedHashes [][]byte

	for _, block := range server.bc.GetBlocksStartingAtHash(highestCommonBlockHash) {
		blockHash := block.GetBlockHeaderHash()
		unsharedHashes = append(unsharedHashes, blockHash)
	}

	requestPeer.SendObjects(INV, objectEntries{
		blockEntries: unsharedHashes,
	})
}

func (server *Server) ReceiveInv(requestPeer *peer, payload objectEntries) {
	var unknownBlockHashes [][]byte
	var unkownTxHashes [][]byte
	for _, txHash := range payload.txEntries {
		tx := server.memoryPool.GetTxWithLock(txHash)
		if tx != nil { //if tx is already known
			continue
		}
		unkownTxHashes = append(unkownTxHashes, txHash)
	}
	for _, blockHash := range payload.blockEntries {
		block := server.bc.GetBlock(blockHash)
		if block != nil { //if block is already known
			continue
		}
		unknownBlockHashes = append(unknownBlockHashes, blockHash)
	}

	requestPeer.SendObjects(GET_DATA, objectEntries{
		txEntries:    unkownTxHashes,
		blockEntries: unknownBlockHashes,
	})
}

func (server *Server) parseBlocks(serializedBlocks [][]byte) (*blockchain.Block, []*blockchain.Block, error) {
	var newBlocks []*blockchain.Block
	var highestKnownBlock *blockchain.Block

	firstBlockBytes := serializedBlocks[0]
	firstRemoteBlock := blockchain.DeserializeBlock(firstBlockBytes)

	highestKnownBlock = server.bc.GetBlock(firstRemoteBlock.Header.PrevBlockHeaderHash)
	highestKnownBlockIdx := -1
	var prevHash []byte
	if highestKnownBlock == nil {
		prevHash = []byte{}
		if !bytes.Equal(firstRemoteBlock.Header.PrevBlockHeaderHash, prevHash) {
			///first remote block's parent isn't known and it isn't a genesis block
			return nil, nil, &blockchain_errors.ErrOrphanBlock{}
		}
	} else {
		prevHash = highestKnownBlock.GetBlockHeaderHash()

	}

	for blockIdx, blockBytes := range serializedBlocks {
		remoteBlock := blockchain.DeserializeBlock(blockBytes)
		if !bytes.Equal(prevHash, remoteBlock.Header.PrevBlockHeaderHash) {
			return nil, nil, &blockchain_errors.ErrOrphanBlock{}
		}
		remoteBlockHash := remoteBlock.GetBlockHeaderHash()
		localBlock := server.bc.GetBlock(remoteBlockHash)
		if localBlock == nil { //block isn't known
			break
		}
		highestKnownBlock = localBlock
		highestKnownBlockIdx = blockIdx
		prevHash = highestKnownBlock.GetBlockHeaderHash()
	}

	if highestKnownBlockIdx == len(serializedBlocks)-1 { //all remote blocks are known
		return highestKnownBlock, nil, nil
	}

	for _, blockBytes := range serializedBlocks[highestKnownBlockIdx+1:] {
		remoteBlock := blockchain.DeserializeBlock(blockBytes)
		if !bytes.Equal(prevHash, remoteBlock.Header.PrevBlockHeaderHash) {
			return nil, nil, &blockchain_errors.ErrOrphanBlock{}
		}
		newBlocks = append(newBlocks, remoteBlock)
		prevHash = remoteBlock.GetBlockHeaderHash()
	}
	return highestKnownBlock, newBlocks, nil
}

func (server *Server) AddBlockToBc(newBlock *blockchain.Block) error {
	err := server.bc.AddBlock(newBlock)
	if err != nil {
		return err
	}

	for _, tx := range newBlock.Transactions {
		//pushes transaction to the front of the mem pool queue (replaces if it's already in the mempool)
		server.memoryPool.PushFrontTxWithLock(tx)
		server.memoryPool.DeleteTxWithLock(tx.Hash())
	}
	return nil
}

func (server *Server) RemoveBlockFromBc(blockHash []byte) error {
	removedBlock := server.bc.GetBlock(blockHash)
	err := server.bc.RemoveBlock(blockHash)
	if err != nil {
		return err
	}
	for _, tx := range removedBlock.Transactions {
		if err = tx.Verify(server.bc.ChainstateDB); err != nil {
			server.memoryPool.DeleteTxsSpendingFromTxUTXOsWithLock(tx)
			server.memoryPool.PushFrontTxWithLock(tx)
		}
	}
	return nil
}

func (server *Server) ReceiveBlocks(requestPeer *peer, serializedBlocks [][]byte) [][]byte {
	highestKnownBlock, newBlocks, err := server.parseBlocks(serializedBlocks)

	if errors.Is(err, &blockchain_errors.ErrOrphanBlock{}) {
		server.SendGetBlocks(requestPeer)
		return nil
	}
	if len(newBlocks) == 0 {
		return nil
	}

	var newChainHeight int
	var highestKnownBlockHash []byte

	if highestKnownBlock != nil { //if any block in the receiving blockchain is known
		newChainHeight = highestKnownBlock.Header.Height + len(newBlocks)
		highestKnownBlockHash = highestKnownBlock.GetBlockHeaderHash()
	} else {
		//receiving blockchain including genesis block
		newChainHeight = len(newBlocks) - 1
		highestKnownBlockHash = []byte{}
	}

	if newChainHeight <= server.bc.Height() {
		return nil
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	if server.bc.GetBlock(highestKnownBlockHash) == nil && !bytes.Equal(highestKnownBlockHash, []byte{}) {
		//no block in the receiving blockchain is known and receiving bc doesn't include genesis block
		return nil
	}

	lastBlockHash := server.bc.LastBlockHash()
	for !bytes.Equal(lastBlockHash, highestKnownBlockHash) {
		err := server.RemoveBlockFromBc(lastBlockHash)
		if err != nil {
			fmt.Printf("Error removing block: %s\n", hex.EncodeToString(lastBlockHash))
			fmt.Println(err.Error())
			//TODO: better error handling
			return nil
		}
		lastBlockHash = server.bc.LastBlockHash()
	}

	var newBlocksHashes [][]byte
	for _, block := range newBlocks {
		err := server.AddBlockToBc(block)
		if err != nil {
			//TODO: better error handling
			fmt.Println("Error adding block")
			fmt.Println(err.Error())
			return nil
		}

		newBlockHash := block.GetBlockHeaderHash()
		newBlocksHashes = append(newBlocksHashes, newBlockHash)
	}

	server.miningChan <- struct{}{}
	return newBlocksHashes
}

func (server *Server) ReceiveTxs(requestPeer *peer, serializedTxs [][]byte) [][]byte {
	var newTxHashes [][]byte
	for _, txBytes := range serializedTxs {
		tx := transactions.Deserialize(txBytes)
		txHash := tx.Hash()
		if server.memoryPool.GetTxWithLock(txHash) != nil {
			continue
		}
		err := server.memoryPool.PushBackTxWithLock(tx)
		if err != nil {
			continue
		}
		newTxHashes = append(newTxHashes, txHash)
	}
	return newTxHashes
}

func (server *Server) ReceiveData(requestPeer *peer, payload objectEntries) {
	var newTxsHashes [][]byte
	var newBlocksHashes [][]byte
	if len(payload.txEntries) > 0 {
		newTxsHashes = server.ReceiveTxs(requestPeer, payload.txEntries)
	}
	if len(payload.blockEntries) > 0 {
		newBlocksHashes = server.ReceiveBlocks(requestPeer, payload.blockEntries)
	}
	server.BroadcastObjects(INV, objectEntries{
		blockEntries: newBlocksHashes,
		txEntries:    newTxsHashes,
	})
}

func (server *Server) ReceiveGetData(requestPeer *peer, payload objectEntries) {
	data := objectEntries{}

	for _, txHash := range payload.txEntries {
		tx := server.memoryPool.GetTxWithLock(txHash)
		if tx == nil {
			continue
		}
		data.txEntries = append(data.txEntries, tx.Serialize())
	}

	for _, blockHash := range payload.blockEntries {
		block := server.bc.GetBlock(blockHash)
		if block == nil {
			continue
		}
		data.blockEntries = append(data.blockEntries, block.Serialize())
	}

	requestPeer.SendObjects(DATA, data)
}

func (server *Server) ReceiveVersion(requestPeer *peer, payload versionPayload) {
	if !payload.ACK {
		requestPeer.sendString("VERSION" + " " + strconv.Itoa(server.bc.Height()) + " " + "ACK")
	} else {
		requestPeer.sendString("VERSION_ACK")
		requestPeer.sendString("GET_ADDR")
		server.ReceiveVersionAck(requestPeer)
	}
	if payload.BestHeight > server.bc.Height() {
		server.SendGetBlocks(requestPeer)
	}
}

func (server *Server) ReceiveVersionAck(requestPeer *peer) {
	fmt.Printf("Connected to peer:%s\n", requestPeer.GetAddress())
	server.peers[requestPeer.GetAddress()] = requestPeer
}

func (server *Server) SendAddresses(requestPeer *peer) {
	var sb strings.Builder
	sb.WriteString("ADDR")
	for currentPeerAddr := range server.peers {
		if currentPeerAddr != requestPeer.GetAddress() {
			sb.WriteString(" " + currentPeerAddr)
		}
	}
	requestPeer.sendString(sb.String())
}

func (server *Server) HandleTcpCommands() {
	for cmd := range server.commands {
		switch cmd.id {
		case GET_ADDR:
			server.SendAddresses(cmd.peer)
		case ADDR:
			server.ReceiveAddresses(ParseAddrsPayload(cmd.args))
		case VERSION:
			server.ReceiveVersion(cmd.peer, ParseVersionPayload(cmd.args))
		case VERSION_ACK:
			server.ReceiveVersionAck(cmd.peer)
		case GET_BLOCKS:
			server.ReceiveGetBlocks(cmd.peer, ParseGetBlocksPayload(cmd.args))
		case INV:
			server.ReceiveInv(cmd.peer, ParseObjects(cmd.args))
		case GET_DATA:
			server.ReceiveGetData(cmd.peer, ParseObjects(cmd.args))
		case DATA:
			server.ReceiveData(cmd.peer, ParseObjects(cmd.args))
		}
	}
}

func (server *Server) ListenForTcpConnections() {
	listener, err := net.Listen("tcp", ":"+BLOCKCHAIN_PORT)
	if err != nil {
		log.Fatalf("unable to start server: %s", err.Error())
	}

	defer listener.Close()
	log.Printf("TCP Server started on :%s", BLOCKCHAIN_PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("failed to accept connection: %s", err.Error())
			continue
		}

		peer := server.NewPeer(conn)
		go peer.ReadInput()
	}
}

func (server *Server) BroadcastObjects(commandID commandID, entries objectEntries) {
	for _, peer := range server.peers {
		peer.SendObjects(commandID, entries)
	}
}
