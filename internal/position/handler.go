package position

import (
	"errors"
	"net/http"

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
	var input CreatePositionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Invalid JSON format",
			},
		})
		return
	}

	if err := h.validator.Struct(&input); err != nil {
		c.JSON(http.StatusUnprocessableEntity, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Validation failed",
				Details: buildValidationDetails(err),
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

	c.JSON(http.StatusCreated, api.Response[*PositionOutput]{
		Data: output,
	})
}

func (h *Handler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrStockNotFound):
		c.JSON(http.StatusNotFound, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "STOCK_NOT_FOUND",
				Message: "Stock not found",
			},
		})
	case errors.Is(err, ErrPositionAlreadyExists):
		c.JSON(http.StatusConflict, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "POSITION_ALREADY_EXISTS",
				Message: "Position already exists for this wallet and stock",
			},
		})
	case errors.Is(err, ErrInvalidPositionQuantity), errors.Is(err, ErrInvalidPositionAveragePrice):
		c.JSON(http.StatusUnprocessableEntity, api.Response[interface{}]{
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

func buildValidationDetails(err error) map[string][]string {
	details := make(map[string][]string)
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			field := e.Field()
			msg := getValidationMessage(e)
			details[field] = append(details[field], msg)
		}
	}
	return details
}

func getValidationMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Minimum value is " + e.Param()
	case "max":
		return "Maximum value is " + e.Param()
	default:
		return "Invalid value"
	}
}
