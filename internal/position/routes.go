package position

import (
	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/middleware"
	"github.com/opinedajr/micro-investing/internal/wallet"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, walletService wallet.Service) {
	wallets := rg.Group("/wallets")
	wallets.Use(middleware.WalletMiddleware(walletService))

	positions := wallets.Group("/:id/positions")
	positions.GET("", h.List)
	positions.GET("/:positionId", h.Find)
	positions.POST("", h.Create)
	positions.PUT("/:positionId", h.Update)
}
