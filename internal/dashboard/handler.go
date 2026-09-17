package dashboard

import (
	"net/http"
	"strconv"

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
		respondInternalServerError(c)
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
		respondInternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, api.Response[*AllocationOutput]{
		Data: output,
	})
}

func (h *Handler) Evolution(c *gin.Context) {
	walletID := c.Param("id")

	input, ok := parseEvolutionInput(c)
	if !ok {
		return
	}

	output, err := h.service.Evolution(c.Request.Context(), walletID, input)
	if err != nil {
		respondInternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, api.Response[*EvolutionOutput]{
		Data: output,
	})
}

func parseEvolutionInput(c *gin.Context) (EvolutionInput, bool) {
	input := EvolutionInput{}
	yearProvided := false

	if yearStr := c.Query("year"); yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil || year <= 0 {
			respondValidationError(c, "Invalid year")
			return input, false
		}
		input.Year = year
		yearProvided = true
	}

	if quarterStr := c.Query("quarter"); quarterStr != "" {
		quarter, err := strconv.Atoi(quarterStr)
		if err != nil || quarter < 1 || quarter > 4 {
			respondValidationError(c, "Invalid quarter")
			return input, false
		}
		if !yearProvided {
			respondValidationError(c, "quarter requires year")
			return input, false
		}
		input.Quarter = quarter
	}

	return input, true
}

func respondValidationError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, api.Response[interface{}]{
		Error: &api.APIError{
			Code:    "VALIDATION_ERROR",
			Message: message,
		},
	})
}

func respondInternalServerError(c *gin.Context) {
	c.JSON(http.StatusInternalServerError, api.Response[interface{}]{
		Error: &api.APIError{
			Code:    "INTERNAL_ERROR",
			Message: "Internal server error",
		},
	})
}
