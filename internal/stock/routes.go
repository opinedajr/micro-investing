package stock

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, h *Handler) {
	stocks := rg.Group("/stocks")
	stocks.GET("", h.List)
	stocks.GET("/:ticker", h.FindByTicker)
}
