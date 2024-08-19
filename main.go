package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/pedrogomes29/blockchain_node/server"
)

func main() {
	minerAddr := flag.String("miner", "", "Miner's wallet address")
	seeds := flag.String("seeds", "", "Comma-separated list of seed addresses")
	flag.Parse()

	// Check if minerAddr is set
	if *minerAddr == "" {
		fmt.Println("Miner's wallet address is required.")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	if *minerAddr == "" {
		fmt.Println("Miner's wallet address is required.")
		fmt.Println("Usage:")
		flag.PrintDefaults()
		os.Exit(1)
	}

	var seedAddresses []string
	if *seeds != "" {
		seedAddresses = strings.Split(*seeds, ",")

		// Regular expression to match "ip:port" format
		ipPortRegex := regexp.MustCompile(`^([a-zA-Z0-9.-]+)$`)

		for _, seed := range seedAddresses {
			if !ipPortRegex.MatchString(seed) {
				fmt.Printf("Invalid seed address format: %s\n", seed)
				fmt.Println("Each seed address must be in the format ip:port (e.g., 192.168.1.1).")
				os.Exit(1)
			}
		}
	}
	
	server := server.NewServer(*minerAddr, seedAddresses)
	server.Run()
}
