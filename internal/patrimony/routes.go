package patrimony

import (
	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/middleware"
	"github.com/opinedajr/micro-investing/internal/wallet"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, walletService wallet.Service) {
	wallets := rg.Group("/wallets")
	wallets.Use(middleware.WalletMiddleware(walletService))

	patrimonies := wallets.Group("/:walletId/patrimonies")
	patrimonies.GET("", h.List)
	patrimonies.POST("", h.Create)
	patrimonies.PUT("/:id", h.Update)

	assets := wallets.Group("/:walletId/assets")
	assets.GET("", h.ListAssets)
	assets.POST("", h.CreateAsset)
	assets.PUT("/:id", h.UpdateAsset)
	assets.DELETE("/:id", h.DeleteAsset)
}
