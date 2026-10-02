package dividends

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/opinedajr/micro-investing/internal/shared/api"
)

type Handler struct {
	service   Service
	validator *validator.Validate
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service:   service,
		validator: validator.New(),
	}
}

func (h *Handler) Create(c *gin.Context) {
	var input CreateDividendInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid JSON format",
			},
		})
		return
	}

	if err := h.validator.Struct(&input); err != nil {
		c.JSON(http.StatusBadRequest, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Validation failed",
			},
		})
		return
	}

	input.WalletID = c.Param("id")

	output, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, api.Response[*DividendOutput]{
		Data: output,
	})
}

func (h *Handler) List(c *gin.Context) {
	filter := DividendFilter{
		WalletID: c.Param("id"),
	}

	if yearStr := c.Query("year"); yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, api.Response[interface{}]{
				Error: &api.APIError{
					Code:    "VALIDATION_ERROR",
					Message: "Invalid year format",
				},
			})
			return
		}
		filter.Year = year
	}

	outputs, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.Response[*DividendsListOutput]{
		Data: &DividendsListOutput{Items: outputs},
	})
}

func (h *Handler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrDividendAlreadyExists):
		c.JSON(http.StatusConflict, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "DIVIDEND_ALREADY_EXISTS",
				Message: "Dividend already exists for this wallet and year",
			},
		})
	case errors.Is(err, ErrInvalidDividendYear), errors.Is(err, ErrInvalidDividendAmount):
		c.JSON(http.StatusBadRequest, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
	default:
		c.JSON(http.StatusInternalServerError, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
			},
		})
	}
}
