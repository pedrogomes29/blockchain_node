package server

import (
	"log"
	"net"
	"strings"
)

const BLOCKCHAIN_PORT string = "8333"

func (server *Server) ConnectToAddress(address string){
	if _,ok := server.peers[address];ok{ //if address is already known
		return;
	}

	conn, err := net.Dial("tcp", address + ":" + BLOCKCHAIN_PORT)
	if err != nil {
		log.Panic("Error establishing connection: ", err)
		return
	}

	newPeer := &peer{
		conn: conn,
		commands: server.commands,
	}

	server.peers[address] = newPeer

	log.Printf("successfully established connection to a new peer: %s\n", address)

	newPeer.send([]byte("GET_ADDR"))
}

func (server *Server) ReceiveAddresses(addressesBytes [][]byte){
	for _, addressBytes := range addressesBytes{
		server.ConnectToAddress(string(addressBytes))
	}
}

func (server *Server) SendAddresses(requestPeer *peer){
	var sb strings.Builder
	sb.WriteString("ADDR")
	for currentPeerAddr := range server.peers{
		if currentPeerAddr != requestPeer.GetAddress() {
			sb.WriteString(" " + currentPeerAddr)
		}
	}
}

func (server *Server) HandleTcpCommands() {
	for cmd := range server.commands {
		switch cmd.id {
			case GET_ADDR:
				server.SendAddresses(cmd.peer)
			case ADDR:
				server.ReceiveAddresses(cmd.args)
			case VERSION:
				//TODO
			case VERSION_ACK:
				//TODO
		}
	}
}

func (server *Server) ListenForTcpConnections(){
	listener, err := net.Listen("tcp", ":" + BLOCKCHAIN_PORT)
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
		server.peers[peer.GetAddress()] = peer
		go peer.ReadInput()
	}
}