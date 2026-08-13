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

func TestHandler_CreateAsset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - creates asset with valid input", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).createAssetFunc = func(ctx context.Context, input CreateAssetInput) (*AssetOutput, error) {
			return &AssetOutput{
				ID:          "asset-id",
				WalletID:    input.WalletID,
				Type:        input.Type,
				Date:        input.Date,
				Description: input.Description,
				Amount:      input.Amount,
				CreatedAt:   "2026-07-25T12:00:00Z",
				UpdatedAt:   "2026-07-25T12:00:00Z",
			}, nil
		}
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-15T12:00:00Z",
			"description": "PETR4 - Petrobras",
			"amount":      150000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/assets", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "walletId", Value: "wallet-id"}}

		handler.CreateAsset(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response api.Response[*AssetOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "asset-id", response.Data.ID)
		assert.Equal(t, "wallet-id", response.Data.WalletID)
		assert.Equal(t, TypeStocks, response.Data.Type)
		assert.Equal(t, int64(150000), response.Data.Amount)
	})

	t.Run("error - returns 422 for invalid json", func(t *testing.T) {
		mockSvc := newMockService(nil)
		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/assets", bytes.NewBuffer([]byte("{invalid")))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "walletId", Value: "wallet-id"}}

		handler.CreateAsset(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 422 for missing fields", func(t *testing.T) {
		mockSvc := newMockService(nil)
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"type": "stocks",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/assets", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "walletId", Value: "wallet-id"}}

		handler.CreateAsset(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 422 for invalid asset type", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).createAssetFunc = func(ctx context.Context, input CreateAssetInput) (*AssetOutput, error) {
			return nil, ErrInvalidAssetType
		}
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"type":        "invalid",
			"date":        "2026-07-15T12:00:00Z",
			"description": "PETR4",
			"amount":      150000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/assets", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "walletId", Value: "wallet-id"}}

		handler.CreateAsset(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 422 for invalid date format", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).createAssetFunc = func(ctx context.Context, input CreateAssetInput) (*AssetOutput, error) {
			return nil, ErrInvalidAssetDate
		}
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"type":        "stocks",
			"date":        "invalid-date",
			"description": "PETR4",
			"amount":      150000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/assets", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "walletId", Value: "wallet-id"}}

		handler.CreateAsset(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 500 for internal server error", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).createAssetFunc = func(ctx context.Context, input CreateAssetInput) (*AssetOutput, error) {
			return nil, errors.New("database error")
		}
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-15T12:00:00Z",
			"description": "PETR4",
			"amount":      150000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/assets", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "walletId", Value: "wallet-id"}}

		handler.CreateAsset(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	})
}

func TestHandler_DeleteAsset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - deletes asset and returns 204", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).deleteAssetFunc = func(ctx context.Context, id string) error {
			return nil
		}
		handler := NewHandler(mockSvc)
		walletService := &mockWalletServiceForRoutes{}

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, walletService)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/v1/wallets/wallet-id/assets/asset-id", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
		assert.Empty(t, w.Body.String())
	})

	t.Run("error - returns 404 when asset not found", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).deleteAssetFunc = func(ctx context.Context, id string) error {
			return ErrAssetNotFound
		}
		handler := NewHandler(mockSvc)
		walletService := &mockWalletServiceForRoutes{}

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, walletService)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/v1/wallets/wallet-id/assets/missing-id", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "ASSET_NOT_FOUND", response.Error.Code)
	})

	t.Run("error - returns 500 for internal server error", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).deleteAssetFunc = func(ctx context.Context, id string) error {
			return errors.New("database error")
		}
		handler := NewHandler(mockSvc)
		walletService := &mockWalletServiceForRoutes{}

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, walletService)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/v1/wallets/wallet-id/assets/asset-id", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	})
}
