package wallet

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	wallets := rg.Group("/wallets")
	wallets.POST("", h.Create)
}
