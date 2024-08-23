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
		tx := server.memoryPool.GetTx(txHash)
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


func (server *Server) parseBlocks(serializedBlocks [][]byte) (*blockchain.Block, []*blockchain.Block,error){
	var newBlocks []*blockchain.Block
	var highestKnownBlock *blockchain.Block


	firstBlockBytes := serializedBlocks[0]
	firstRemoteBlock := blockchain.DeserializeBlock(firstBlockBytes)
	highestKnownBlock = server.bc.GetBlock(firstRemoteBlock.Header.PrevBlockHeaderHash)
	if(highestKnownBlock==nil){ //first remote block's parent isn't known 
		return nil, nil, &blockchain_errors.ErrOrphanBlock{}
	}
	highestKnownBlockIdx := -1

	prevHash := highestKnownBlock.GetBlockHeaderHash()
	for blockIdx, blockBytes := range serializedBlocks {
		remoteBlock := blockchain.DeserializeBlock(blockBytes)
		if(!bytes.Equal(prevHash,remoteBlock.Header.PrevBlockHeaderHash)){
			return nil, nil, &blockchain_errors.ErrOrphanBlock{}
		}
		remoteBlockHash := remoteBlock.GetBlockHeaderHash()
		localBlock := server.bc.GetBlock(remoteBlockHash)
		if localBlock==nil{ //block isn't known
			break
		}
		highestKnownBlock = localBlock
		highestKnownBlockIdx = blockIdx
		prevHash = highestKnownBlock.GetBlockHeaderHash()
	}

	if highestKnownBlockIdx == len(serializedBlocks)-1{ //all remote blocks are known
		return highestKnownBlock,nil,nil
	}

	for _, blockBytes := range serializedBlocks[highestKnownBlockIdx+1:] {
		remoteBlock := blockchain.DeserializeBlock(blockBytes)
		if(!bytes.Equal(prevHash,remoteBlock.Header.PrevBlockHeaderHash)){
			return nil, nil, &blockchain_errors.ErrOrphanBlock{}
		}
		newBlocks = append(newBlocks, remoteBlock)
		prevHash = remoteBlock.GetBlockHeaderHash()
	}
	return highestKnownBlock, newBlocks, nil
}

func (server *Server) ReceiveBlocks(requestPeer *peer, serializedBlocks [][]byte) [][]byte{
	highestKnownBlock, newBlocks, err := server.parseBlocks(serializedBlocks)
	if errors.Is(err, &blockchain_errors.ErrOrphanBlock{}) {
		server.SendGetBlocks(requestPeer)
		return nil
	}
	if len(newBlocks)==0{
		return nil
	}
	newChainHeight := highestKnownBlock.Header.Height + len(newBlocks) 
	if newChainHeight <= server.bc.Height(){
		return nil
	}

	server.mu.Lock()
	defer server.mu.Unlock()

	for !bytes.Equal(server.bc.LastBlockHash(),highestKnownBlock.GetBlockHeaderHash()){
		//TODO: remove blocks after highest known
	}

	var newBlocksHashes [][]byte
	for _, block := range newBlocks {
		err := server.bc.AddBlock(block)
		if err != nil {
			//TODO: better error handling
			return nil
		}

		newBlockHash := block.GetBlockHeaderHash()
		newBlocksHashes = append(newBlocksHashes, newBlockHash)
	}

	server.miningChan <- struct{}{}
	return newBlocksHashes
}

func (server *Server) ReceiveData(requestPeer *peer, payload objectEntries) {
	//TODO: Receive TXs
	newBlocksHashes := server.ReceiveBlocks(requestPeer, payload.blockEntries)
	server.BroadcastObjects(INV, objectEntries{
		blockEntries: newBlocksHashes,
	})
}

func (server *Server) ReceiveGetData(requestPeer *peer, payload objectEntries) {
	data := objectEntries{}

	for _, txHash := range payload.txEntries {
		tx := server.memoryPool.GetTx(txHash)
		if tx == nil {
			//TODO: Error handling in case tx isn't in memory pool anymore
		}
		data.txEntries = append(data.txEntries, tx.Serialize())
	}

	for _, blockHash := range payload.blockEntries {
		block := server.bc.GetBlock(blockHash)
		if block == nil {
			//TODO: Error handling in case block isn't in local blockchain anymore
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
