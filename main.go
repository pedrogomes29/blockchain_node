package main

import (
	"fmt"
	"time"

	"github.com/pedrogomes29/blockchain/server"
	"github.com/pedrogomes29/blockchain/wallet"
)

func main() {
	// Create two wallets
	minerWallet := wallet.NewWalletAndPrivateKey()

	wallet1 := wallet.NewWalletAndPrivateKey()

	wallet2 := wallet.NewWalletAndPrivateKey()

	// Start a server with the first wallet's address
	server := server.NewServer(minerWallet.Address())
	go server.Run()

	minerWallet.SetServer(server)
	wallet1.SetServer(server)
	wallet2.SetServer(server)


	time.Sleep(3*time.Second)

	minerWallet.SendToAddress(wallet1.Address(), 10)

	time.Sleep(3*time.Second)

	// Initial balances
	fmt.Printf("Initial balance of %s: %d\n", wallet1.Address(), wallet1.GetBalance())
	fmt.Printf("Initial balance of %s: %d\n", wallet2.Address(), wallet2.GetBalance())

	// Perform some transactions
	fmt.Println("Performing transactions...")

	wallet1.SendToAddress(wallet2.Address(),10)
	
	// Send 10 coins from wallet1 to wallet2
	time.Sleep(3*time.Second)

	// Check balances after the first transaction
	fmt.Printf("Balance of %s after sending 10 coins: %d\n", wallet1.Address(), wallet1.GetBalance())
	fmt.Printf("Balance of %s after receiving 10 coins: %d\n", wallet2.Address(), wallet2.GetBalance())

	// Send 5 coins back from wallet2 to wallet1
	wallet2.SendToAddress(wallet1.Address(), 5)
	time.Sleep(3*time.Second)

	// Check balances after the second transaction
	fmt.Printf("Balance of %s after receiving 5 coins: %d\n", wallet1.Address(), wallet1.GetBalance())
	fmt.Printf("Balance of %s after sending 5 coins: %d\n", wallet2.Address(), wallet2.GetBalance())

	// Send 2 coins from wallet1 to wallet2
	wallet1.SendToAddress(wallet2.Address(), 2)
	time.Sleep(3*time.Second)

	// Final balances
	fmt.Printf("Balance of %s after sending 2 coins: %d\n", wallet1.Address(), wallet1.GetBalance())
	fmt.Printf("Balance of %s after receiving 3 coins: %d\n", wallet2.Address(), wallet2.GetBalance())
	
}
