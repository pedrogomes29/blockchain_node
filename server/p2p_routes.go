package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (server *Server) AddP2PRoutes(r *gin.Engine) {
	p2pRoutes := r.Group("/p2p")
	{
		p2pRoutes.GET("/", func(c *gin.Context) {
			c.JSON(http.StatusOK, "OK")
		})
	}
}
