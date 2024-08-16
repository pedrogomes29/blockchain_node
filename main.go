package main

import (
	"flag"

	"github.com/pedrogomes29/blockchain/server"
)

func main() {
	minerAddr := flag.String("miner", "", "Miner's wallet address")
	flag.Parse()

	server := server.NewServer(*minerAddr)
	server.Run()
}
