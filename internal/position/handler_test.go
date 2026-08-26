package position

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/opinedajr/micro-investing/internal/shared/api"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - creates position with derived values", func(t *testing.T) {
		mockSvc := newMockServiceWithCreate(func(ctx context.Context, input CreatePositionInput) (*PositionOutput, error) {
			return &PositionOutput{
				ID:               "position-id",
				WalletID:         input.WalletID,
				StockID:          input.StockID,
				Quantity:         input.Quantity,
				AveragePrice:     input.AveragePrice,
				CurrentPrice:     7500,
				Invested:         500000,
				Balance:          750000,
				VariationValue:   250000,
				VariationPercent: 50,
				PortfolioPercent: 100,
				CreatedAt:        "2026-08-26T12:00:00Z",
				UpdatedAt:        "2026-08-26T12:00:00Z",
			}, nil
		})

		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"stock_id":      "stock-id",
			"quantity":      100,
			"average_price": 5000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/positions", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response api.Response[*PositionOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "position-id", response.Data.ID)
		assert.Equal(t, "wallet-id", response.Data.WalletID)
		assert.Equal(t, int64(7500), response.Data.CurrentPrice)
		assert.Equal(t, int64(500000), response.Data.Invested)
	})

	t.Run("error - returns 422 for invalid json", func(t *testing.T) {
		handler := NewHandler(newMockServiceWithCreate(nil))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/positions", bytes.NewBuffer([]byte("{invalid")))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 422 for validation error", func(t *testing.T) {
		handler := NewHandler(newMockServiceWithCreate(nil))

		body := map[string]interface{}{"stock_id": "stock-id"}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/positions", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 404 when stock not found", func(t *testing.T) {
		mockSvc := newMockServiceWithCreate(func(ctx context.Context, input CreatePositionInput) (*PositionOutput, error) {
			return nil, ErrStockNotFound
		})
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"stock_id":      "missing-stock",
			"quantity":      10,
			"average_price": 1000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/positions", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "STOCK_NOT_FOUND", response.Error.Code)
	})

	t.Run("error - returns 409 when position already exists", func(t *testing.T) {
		mockSvc := newMockServiceWithCreate(func(ctx context.Context, input CreatePositionInput) (*PositionOutput, error) {
			return nil, ErrPositionAlreadyExists
		})
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"stock_id":      "stock-id",
			"quantity":      10,
			"average_price": 1000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/positions", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "POSITION_ALREADY_EXISTS", response.Error.Code)
	})

	t.Run("error - returns 422 for domain validation error", func(t *testing.T) {
		mockSvc := newMockServiceWithCreate(func(ctx context.Context, input CreatePositionInput) (*PositionOutput, error) {
			return nil, ErrInvalidPositionQuantity
		})
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"stock_id":      "stock-id",
			"quantity":      0,
			"average_price": 1000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/positions", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 500 for internal server error", func(t *testing.T) {
		mockSvc := newMockServiceWithCreate(func(ctx context.Context, input CreatePositionInput) (*PositionOutput, error) {
			return nil, errors.New("database error")
		})
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"stock_id":      "stock-id",
			"quantity":      10,
			"average_price": 1000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/positions", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	})
}

func TestHandler_buildValidationDetails(t *testing.T) {
	v := validator.New()
	type input struct {
		Quantity int64 `json:"quantity" validate:"required,min=1"`
	}

	err := v.Struct(&input{Quantity: 0})
	details := buildValidationDetails(err)

	assert.NotEmpty(t, details)
	assert.Contains(t, details, "Quantity")
}
