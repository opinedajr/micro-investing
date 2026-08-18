package stock

import (
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup, handler Handler) {
	router.GET("/stocks", handler.ListStocks)
	router.GET("/stocks/:ticker", handler.GetStockByTicker)
}
