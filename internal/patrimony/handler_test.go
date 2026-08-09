package patrimony

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/api"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - creates patrimony with valid input", func(t *testing.T) {
		mockSvc := newMockService(func(ctx context.Context, input CreatePatrimonyInput) (*PatrimonyOutput, error) {
			return &PatrimonyOutput{
				ID:        "patrimony-id",
				WalletID:  input.WalletID,
				Year:      input.Year,
				Month:     input.Month,
				Type:      input.Type,
				Amount:    input.Amount,
				CreatedAt: "2026-07-25T12:00:00Z",
				UpdatedAt: "2026-07-25T12:00:00Z",
			}, nil
		})

		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"year":   2026,
			"month":  7,
			"type":   "stocks",
			"amount": 150000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/patrimonies", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "walletId", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response api.Response[*PatrimonyOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "patrimony-id", response.Data.ID)
		assert.Equal(t, "wallet-id", response.Data.WalletID)
		assert.Equal(t, 2026, response.Data.Year)
		assert.Equal(t, 7, response.Data.Month)
		assert.Equal(t, TypeStocks, response.Data.Type)
		assert.Equal(t, int64(150000), response.Data.Amount)
	})

	t.Run("error - returns 422 for invalid json", func(t *testing.T) {
		mockSvc := newMockService(nil)
		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/patrimonies", bytes.NewBuffer([]byte("{invalid")))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "walletId", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 422 for validation error", func(t *testing.T) {
		mockSvc := newMockService(nil)
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"year": 2026,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/patrimonies", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "walletId", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 409 when patrimony already exists", func(t *testing.T) {
		mockSvc := newMockService(func(ctx context.Context, input CreatePatrimonyInput) (*PatrimonyOutput, error) {
			return nil, ErrPatrimonyAlreadyExists
		})
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"year":   2026,
			"month":  7,
			"type":   "stocks",
			"amount": 150000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/patrimonies", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "walletId", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "PATRIMONY_ALREADY_EXISTS", response.Error.Code)
	})

	t.Run("error - returns 422 for domain validation error", func(t *testing.T) {
		mockSvc := newMockService(func(ctx context.Context, input CreatePatrimonyInput) (*PatrimonyOutput, error) {
			return nil, ErrInvalidAssetType
		})
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"year":   2026,
			"month":  7,
			"type":   "stocks",
			"amount": 150000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/patrimonies", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "walletId", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 500 for internal server error", func(t *testing.T) {
		mockSvc := newMockService(func(ctx context.Context, input CreatePatrimonyInput) (*PatrimonyOutput, error) {
			return nil, errors.New("database connection failed")
		})
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"year":   2026,
			"month":  7,
			"type":   "stocks",
			"amount": 150000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/patrimonies", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "walletId", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	})
}

func TestHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - updates patrimony", func(t *testing.T) {
		mockSvc := newMockServiceWithUpdate(func(ctx context.Context, id string, input UpdatePatrimonyInput) (*PatrimonyOutput, error) {
			return &PatrimonyOutput{
				ID:        id,
				WalletID:  input.WalletID,
				Year:      input.Year,
				Month:     input.Month,
				Type:      input.Type,
				Amount:    input.Amount,
				CreatedAt: "2026-07-25T10:00:00Z",
				UpdatedAt: "2026-07-25T14:00:00Z",
			}, nil
		})
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"year":   2026,
			"month":  7,
			"type":   "stocks",
			"amount": 200000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/api/v1/wallets/wallet-id/patrimonies/patrimony-id", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{
			{Key: "walletId", Value: "wallet-id"},
			{Key: "id", Value: "patrimony-id"},
		}

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[*PatrimonyOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "patrimony-id", response.Data.ID)
		assert.Equal(t, int64(200000), response.Data.Amount)
	})

	t.Run("error - returns 422 for invalid update input", func(t *testing.T) {
		mockSvc := newMockServiceWithUpdate(nil)
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"year": 2026,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/api/v1/wallets/wallet-id/patrimonies/patrimony-id", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{
			{Key: "walletId", Value: "wallet-id"},
			{Key: "id", Value: "patrimony-id"},
		}

		handler.Update(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 404 when patrimony not found", func(t *testing.T) {
		mockSvc := newMockServiceWithUpdate(func(ctx context.Context, id string, input UpdatePatrimonyInput) (*PatrimonyOutput, error) {
			return nil, ErrPatrimonyNotFound
		})
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"year":   2026,
			"month":  7,
			"type":   "stocks",
			"amount": 200000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/api/v1/wallets/wallet-id/patrimonies/patrimony-id", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{
			{Key: "walletId", Value: "wallet-id"},
			{Key: "id", Value: "patrimony-id"},
		}

		handler.Update(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "PATRIMONY_NOT_FOUND", response.Error.Code)
	})

	t.Run("error - returns 409 when patrimony already exists", func(t *testing.T) {
		mockSvc := newMockServiceWithUpdate(func(ctx context.Context, id string, input UpdatePatrimonyInput) (*PatrimonyOutput, error) {
			return nil, ErrPatrimonyAlreadyExists
		})
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"year":   2026,
			"month":  7,
			"type":   "stocks",
			"amount": 200000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/api/v1/wallets/wallet-id/patrimonies/patrimony-id", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{
			{Key: "walletId", Value: "wallet-id"},
			{Key: "id", Value: "patrimony-id"},
		}

		handler.Update(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "PATRIMONY_ALREADY_EXISTS", response.Error.Code)
	})
}

func TestHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - lists patrimonies with filters", func(t *testing.T) {
		mockSvc := newMockServiceWithList(func(ctx context.Context, filter PatrimonyFilter) ([]PatrimonyOutput, error) {
			return []PatrimonyOutput{
				{
					ID:        "patrimony-1",
					WalletID:  filter.WalletID,
					Year:      2026,
					Month:     7,
					Type:      TypeStocks,
					Amount:    100000,
					CreatedAt: "2026-07-25T10:00:00Z",
					UpdatedAt: "2026-07-25T10:00:00Z",
				},
			}, nil
		})
		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/patrimonies?type=stocks&year=2026&month=7", nil)
		c.Params = []gin.Param{{Key: "walletId", Value: "wallet-id"}}

		handler.List(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[[]PatrimonyOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, "patrimony-1", response.Data[0].ID)
	})

	t.Run("error - returns 500 for internal server error", func(t *testing.T) {
		mockSvc := newMockServiceWithList(func(ctx context.Context, filter PatrimonyFilter) ([]PatrimonyOutput, error) {
			return nil, errors.New("database error")
		})
		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/patrimonies", nil)
		c.Params = []gin.Param{{Key: "walletId", Value: "wallet-id"}}

		handler.List(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	})
}
