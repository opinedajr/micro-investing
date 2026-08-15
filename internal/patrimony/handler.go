package patrimony

import (
	"errors"
	"net/http"
	"strconv"
	"time"

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

func (h *Handler) CreateAsset(c *gin.Context) {
	var input CreateAssetInput
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

	output, err := h.service.CreateAsset(c.Request.Context(), input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, api.Response[*AssetOutput]{
		Data: output,
	})
}

func (h *Handler) UpdateAsset(c *gin.Context) {
	var input UpdateAssetInput
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

	output, err := h.service.UpdateAsset(c.Request.Context(), id, input)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.Response[*AssetOutput]{
		Data: output,
	})
}

func (h *Handler) DeleteAsset(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteAsset(c.Request.Context(), id); err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListAssets(c *gin.Context) {
	filter := AssetFilter{
		WalletID: c.Param("walletId"),
	}

	if assetType := c.Query("type"); assetType != "" {
		filter.Type = AssetType(assetType)
	}
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, api.Response[interface{}]{
				Error: &api.APIError{
					Code:    "VALIDATION_ERROR",
					Message: "Invalid start_date format, expected YYYY-MM-DD",
				},
			})
			return
		}
		filter.StartDate = &startDate
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, api.Response[interface{}]{
				Error: &api.APIError{
					Code:    "VALIDATION_ERROR",
					Message: "Invalid end_date format, expected YYYY-MM-DD",
				},
			})
			return
		}
		filter.EndDate = &endDate
	}

	outputs, err := h.service.ListAssets(c.Request.Context(), filter)
	if err != nil {
		h.handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, api.Response[[]AssetOutput]{
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
	case errors.Is(err, ErrAssetNotFound):
		c.JSON(http.StatusNotFound, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "ASSET_NOT_FOUND",
				Message: "Asset not found",
			},
		})
	case errors.Is(err, ErrInvalidAssetType), errors.Is(err, ErrInvalidPatrimonyYear), errors.Is(err, ErrInvalidPatrimonyMonth), errors.Is(err, ErrInvalidPatrimonyAmount), errors.Is(err, ErrInvalidAssetDate), errors.Is(err, ErrInvalidAssetDescription), errors.Is(err, ErrInvalidAssetAmount), errors.Is(err, ErrInvalidDateRange):
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
