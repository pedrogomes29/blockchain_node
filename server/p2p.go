package server

import (
	"log"
	"net"
	"strconv"
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

	log.Printf("successfully established connection to a new peer: %s\n", address)

	newPeer.sendString("VERSION" + " " + strconv.Itoa(server.bc.Height))

	go newPeer.ReadInput()
}

func (server *Server) ReceiveAddresses(addressesBytes [][]byte){
	for _, addressBytes := range addressesBytes{
		server.ConnectToAddress(string(addressBytes))
	}
}


func (server *Server) ReceiveVersion(requestPeer *peer, addressesBytes [][]byte){
	_,err := strconv.Atoi(string(addressesBytes[0])) //TODO: react to received version
	if err!=nil{
		log.Panicf("error parsing peer's blockchain height %s", string(addressesBytes[0]))
	}
	requestPeer.sendString("VERSION_ACK" + " " + strconv.Itoa(server.bc.Height))
}

func (server *Server) ReceiveVersionAck(requestPeer *peer, addressesBytes [][]byte){
	server.peers[requestPeer.GetAddress()] = requestPeer
	piggyBackedVersion := len(addressesBytes)>0
	if piggyBackedVersion {
		_,err := strconv.Atoi(string(addressesBytes[0]))
		//TODO: react to received version
		if err!=nil{
			log.Panicf("error parsing peer's blockchain height %s", string(addressesBytes[0]))
		}
		requestPeer.sendString("VERSION_ACK")
		requestPeer.sendString("GET_ADDR")
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
	requestPeer.sendString(sb.String())
}

func (server *Server) HandleTcpCommands() {
	for cmd := range server.commands {
		switch cmd.id {
			case GET_ADDR:
				server.SendAddresses(cmd.peer)
			case ADDR:
				server.ReceiveAddresses(cmd.args)
			case VERSION:
				server.ReceiveVersion(cmd.peer, cmd.args)
			case VERSION_ACK:
				server.ReceiveVersionAck(cmd.peer, cmd.args)
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
		go peer.ReadInput()
	}
}