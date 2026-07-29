package wallet

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/opinedajr/micro-investing/internal/shared"
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
	var input CreateWalletInput
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
		var details map[string][]string
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			details = make(map[string][]string)
			for _, e := range validationErrors {
				field := e.Field()
				msg := getValidationMessage(e)
				details[field] = append(details[field], msg)
			}
		}

		c.JSON(http.StatusUnprocessableEntity, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Validation failed",
				Details: details,
			},
		})
		return
	}

	input.UserID = shared.DefaultUserID

	output, err := h.service.Create(c.Request.Context(), input)
	if err != nil {
		if err == ErrWalletNameAlreadyExists {
			c.JSON(http.StatusConflict, api.Response[interface{}]{
				Error: &api.APIError{
					Code:    "WALLET_NAME_ALREADY_EXISTS",
					Message: "Wallet name already exists",
				},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
			},
		})
		return
	}

	c.JSON(http.StatusCreated, api.Response[*WalletOutput]{
		Data: output,
	})
}

func (h *Handler) List(c *gin.Context) {
	outputs, err := h.service.List(c.Request.Context(), shared.DefaultUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Internal server error",
			},
		})
		return
	}

	c.JSON(http.StatusOK, api.Response[[]WalletOutput]{
		Data: outputs,
		Meta: nil,
	})
}

func (h *Handler) Find(c *gin.Context) {
	id := c.Param("id")

	output, err := h.service.Find(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, api.Response[interface{}]{
			Error: &api.APIError{
				Code:    "WALLET_NOT_FOUND",
				Message: "Wallet not found",
			},
		})
		return
	}

	c.JSON(http.StatusOK, api.Response[*WalletOutput]{
		Data: output,
	})
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
