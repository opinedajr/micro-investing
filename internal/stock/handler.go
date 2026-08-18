package stock

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/api"
)

type handler struct {
	service Service
}

func NewHandler(service Service) Handler {
	return &handler{service: service}
}

func (h *handler) ListStocks(c *gin.Context) {
	stocks, err := h.service.ListStocks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	output := make([]StockOutput, len(stocks))
	for i, stock := range stocks {
		output[i] = StockOutput{
			ID:      stock.ID,
			Ticker:  stock.Ticker,
			Name:    stock.Name,
			Sector:  stock.Sector,
			Rank:    stock.Rank,
			Website: stock.Website,
		}
	}

	c.JSON(http.StatusOK, ListOutput{Stocks: output})
}

func (h *handler) GetStockByTicker(c *gin.Context) {
	ticker := c.Param("ticker")

	stock, err := h.service.GetStockByTicker(c.Request.Context(), ticker)
	if err != nil {
		if err == ErrStockNotFound {
			c.JSON(http.StatusNotFound, api.Response[interface{}]{
				Error: &api.APIError{
					Code:    "STOCK_NOT_FOUND",
					Message: err.Error(),
				},
			})
			return
		}
		c.JSON(http.StatusInternalServerError, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	output := StockOutput{
		ID:      stock.ID,
		Ticker:  stock.Ticker,
		Name:    stock.Name,
		Sector:  stock.Sector,
		Rank:    stock.Rank,
		Website: stock.Website,
	}

	c.JSON(http.StatusOK, output)
}
