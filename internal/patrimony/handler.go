package patrimony

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
	var input CreatePatrimonyInput
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

	input.WalletID = c.Param("walletId")

	output, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, api.Response[*PatrimonyOutput]{
		Data: output,
	})
}

func (h *Handler) Update(c *gin.Context) {
	var input UpdatePatrimonyInput
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

	input.WalletID = c.Param("walletId")
	id := c.Param("id")

	output, err := h.service.Update(c.Request.Context(), id, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.Response[*PatrimonyOutput]{
		Data: output,
	})
}

func (h *Handler) List(c *gin.Context) {
	filter := PatrimonyFilter{
		WalletID: c.Param("walletId"),
	}

	if assetType := c.Query("type"); assetType != "" {
		filter.Type = AssetType(assetType)
	}
	if yearStr := c.Query("year"); yearStr != "" {
		year, err := strconv.Atoi(yearStr)
		if err == nil {
			filter.Year = year
		}
	}
	if monthStr := c.Query("month"); monthStr != "" {
		month, err := strconv.Atoi(monthStr)
		if err == nil {
			filter.Month = month
		}
	}

	outputs, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, api.Response[[]PatrimonyOutput]{
		Data: outputs,
	})
}

func (h *Handler) handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrPatrimonyAlreadyExists):
		c.JSON(http.StatusConflict, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "PATRIMONY_ALREADY_EXISTS",
				Message: "Patrimony already exists for this wallet, year, month and type",
			},
		})
	case errors.Is(err, ErrPatrimonyNotFound):
		c.JSON(http.StatusNotFound, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "PATRIMONY_NOT_FOUND",
				Message: "Patrimony not found",
			},
		})
	case errors.Is(err, ErrInvalidAssetType), errors.Is(err, ErrInvalidPatrimonyYear), errors.Is(err, ErrInvalidPatrimonyMonth), errors.Is(err, ErrInvalidPatrimonyAmount):
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
		return "Minimum length is " + e.Param()
	case "max":
		return "Maximum length is " + e.Param()
	default:
		return "Invalid value"
	}
}
