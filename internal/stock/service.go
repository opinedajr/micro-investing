package stock

import (
	"context"

	"github.com/gin-gonic/gin"
)

type Service interface {
	ListStocks(ctx context.Context) ([]Stock, error)
	GetStockByTicker(ctx context.Context, ticker string) (*Stock, error)
	SeedStocks(ctx context.Context, force bool) (inserted, updated, skipped int, err error)
}

type Handler interface {
	ListStocks(c *gin.Context)
	GetStockByTicker(c *gin.Context)
}
