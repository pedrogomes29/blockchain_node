package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/pedrogomes29/blockchain_node/server"
)

func main() {
	minerAddr := flag.String("miner", "", "Miner's wallet address")
	flag.Parse()

	// Check if minerAddr is set
	if *minerAddr == "" {
		fmt.Println("Miner's wallet address is required.")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Start the server with the provided miner's address
	server := server.NewServer(*minerAddr)
	server.Run()
}
