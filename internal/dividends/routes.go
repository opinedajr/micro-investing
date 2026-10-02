package dividends

import (
	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/middleware"
	"github.com/opinedajr/micro-investing/internal/wallet"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, walletService wallet.Service) {
	wallets := rg.Group("/wallets")
	wallets.Use(middleware.WalletMiddleware(walletService))

	dividends := wallets.Group("/:id/dividends")
	dividends.GET("", h.List)
	dividends.POST("", h.Create)
	dividends.PUT("/:dividendId", h.Update)
}
