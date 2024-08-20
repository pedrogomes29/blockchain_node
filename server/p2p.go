package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"strings"
)

const BLOCKCHAIN_PORT string = "8333"

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

	log.Printf("successfully established connection to a new peer: %s\n", address)

	fmt.Printf("\nSent height %d\n\n", server.bc.Height)
	newPeer.sendString("VERSION" + " " + strconv.Itoa(server.bc.Height))

	go newPeer.ReadInput()
}

func (server *Server) ReceiveAddresses(addresses addrPayload) {
	for _, address := range addresses {
		server.ConnectToAddress(address)
	}
}

func (server *Server) ReceiveVersion(requestPeer *peer, payload versionPayload) {
	fmt.Printf("\nReceived height %d\n\n", payload.BestHeight)
	if !payload.ACK {
		fmt.Printf("\nSent height %d\n\n", server.bc.Height)
		requestPeer.sendString("VERSION" + " " + strconv.Itoa(server.bc.Height) + " " + "ACK")
	} else {
		requestPeer.sendString("VERSION_ACK")
		requestPeer.sendString("GET_ADDR")
		server.ReceiveVersionAck(requestPeer)
	}
}

func (server *Server) ReceiveVersionAck(requestPeer *peer) {
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
