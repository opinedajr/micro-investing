package stock

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/api"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(c *gin.Context) {
	outputs, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, api.Response[[]StockOutput]{
		Data: outputs,
	})
}

func (h *Handler) FindByTicker(c *gin.Context) {
	ticker := c.Param("ticker")

	output, err := h.service.FindByTicker(c.Request.Context(), ticker)
	if err != nil {
		c.JSON(http.StatusNotFound, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "STOCK_NOT_FOUND",
				Message: "Stock not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, api.Response[*StockOutput]{
		Data: output,
	})
}
