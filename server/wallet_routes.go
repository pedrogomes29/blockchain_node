package server

import (
	"encoding/hex"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/pedrogomes29/blockchain_node/blockchain_errors"
	"github.com/pedrogomes29/blockchain_node/transactions"
)

func (server *Server) AddTransactionHandler(c *gin.Context) {
	var tx transactions.Transaction

	if err := c.ShouldBindJSON(&tx); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid transaction format"})
		return
	}

	if err := server.AddTxToMemPool(tx); err != nil {
		if errors.Is(err, &blockchain_errors.ErrInvalidTxInputSignature{}) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Transaction inputs have at least one invalid signature"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "An unexpected error ocurred"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "Transaction added to mempool successfully"})
}

func (server *Server) FindUTXOsHandler(c *gin.Context) {
	pubKeyHashStr := c.Query("pubKeyHash")
	pubKeyHash, err := hex.DecodeString(pubKeyHashStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid public key hash format"})
		return
	}

	utxos, err := server.FindUTXOs(pubKeyHash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding UTXOs"})
		return
	}

	c.JSON(http.StatusOK, utxos)
}

func (server *Server) FindSpendableUTXOsHandler(c *gin.Context) {
	pubKeyHashStr := c.Query("pubKeyHash")
	amountStr := c.Query("amount")

	pubKeyHash, err := hex.DecodeString(pubKeyHashStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid public key hash format"})
		return
	}

	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount format"})
		return
	}

	utxosTotal, spendableUTXOs, err := server.FindSpendableUTXOs(pubKeyHash, amount)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding spendable UTXOs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total":     utxosTotal,
		"spendable": spendableUTXOs,
	})
}

func (server *Server) AddWalletRoutes(r *gin.Engine) {
	walletRoutes := r.Group("/wallet")
	{
		walletRoutes.POST("/transactions", server.AddTransactionHandler)
		walletRoutes.GET("/utxos", server.FindUTXOsHandler)
		walletRoutes.GET("/spendable_utxos", server.FindSpendableUTXOsHandler)
	}
}
