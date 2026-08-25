package patrimony

import (
	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/middleware"
	"github.com/opinedajr/micro-investing/internal/wallet"
)

func RegisterRoutes(rg *gin.RouterGroup, h *Handler, walletService wallet.Service) {
	wallets := rg.Group("/wallets")
	wallets.Use(middleware.WalletMiddleware(walletService))

	patrimonies := wallets.Group("/:id/patrimonies")
	patrimonies.GET("", h.List)
	patrimonies.POST("", h.Create)
	patrimonies.PUT("/:patrimonyId", h.Update)

	assets := wallets.Group("/:id/assets")
	assets.GET("", h.ListAssets)
	assets.POST("", h.CreateAsset)
	assets.PUT("/:assetId", h.UpdateAsset)
	assets.DELETE("/:assetId", h.DeleteAsset)
}
