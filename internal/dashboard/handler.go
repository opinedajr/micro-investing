package dashboard

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

func (h *Handler) Summary(c *gin.Context) {
	walletID := c.Param("id")

	output, err := h.service.Summary(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, api.Response[*SummaryOutput]{
		Data: output,
	})
}

func (h *Handler) Allocation(c *gin.Context) {
	walletID := c.Param("id")

	output, err := h.service.Allocation(c.Request.Context(), walletID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, api.Response[*AllocationOutput]{
		Data: output,
	})
}
