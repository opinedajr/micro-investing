package patrimony

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/api"
	"github.com/stretchr/testify/assert"
)

func TestHandler_UpdateAsset(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - updates asset with valid input", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).updateAssetFunc = func(ctx context.Context, id string, input UpdateAssetInput) (*AssetOutput, error) {
			return &AssetOutput{
				ID:          id,
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
			"date":        "2026-07-20T12:00:00Z",
			"description": "Updated description",
			"amount":      200000,
		}
		jsonBody, _ := json.Marshal(body)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, &mockWalletServiceForRoutes{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/wallets/wallet-id/assets/asset-id", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[*AssetOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "asset-id", response.Data.ID)
		assert.Equal(t, "wallet-id", response.Data.WalletID)
		assert.Equal(t, TypeStocks, response.Data.Type)
		assert.Equal(t, int64(200000), response.Data.Amount)
		assert.Equal(t, "Updated description", response.Data.Description)
	})

	t.Run("error - returns 422 for invalid json", func(t *testing.T) {
		mockSvc := newMockService(nil)
		handler := NewHandler(mockSvc)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, &mockWalletServiceForRoutes{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/wallets/wallet-id/assets/asset-id", bytes.NewBuffer([]byte("{invalid")))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

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

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, &mockWalletServiceForRoutes{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/wallets/wallet-id/assets/asset-id", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 404 when asset not found", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).updateAssetFunc = func(ctx context.Context, id string, input UpdateAssetInput) (*AssetOutput, error) {
			return nil, ErrAssetNotFound
		}
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-20T12:00:00Z",
			"description": "Updated description",
			"amount":      200000,
		}
		jsonBody, _ := json.Marshal(body)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, &mockWalletServiceForRoutes{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/wallets/wallet-id/assets/missing-id", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "ASSET_NOT_FOUND", response.Error.Code)
	})

	t.Run("error - returns 422 for future date", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).updateAssetFunc = func(ctx context.Context, id string, input UpdateAssetInput) (*AssetOutput, error) {
			return nil, ErrInvalidAssetDate
		}
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-20T12:00:00Z",
			"description": "Updated description",
			"amount":      200000,
		}
		jsonBody, _ := json.Marshal(body)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, &mockWalletServiceForRoutes{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/wallets/wallet-id/assets/asset-id", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 500 for internal server error", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).updateAssetFunc = func(ctx context.Context, id string, input UpdateAssetInput) (*AssetOutput, error) {
			return nil, errors.New("database error")
		}
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-20T12:00:00Z",
			"description": "Updated description",
			"amount":      200000,
		}
		jsonBody, _ := json.Marshal(body)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, &mockWalletServiceForRoutes{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/wallets/wallet-id/assets/asset-id", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	})
}

func TestHandler_ListAssets(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - lists assets for wallet with no filters", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).listAssetsFunc = func(ctx context.Context, filter AssetFilter) ([]AssetOutput, error) {
			assert.Equal(t, "wallet-id", filter.WalletID)
			assert.Empty(t, filter.Type)
			assert.Nil(t, filter.StartDate)
			assert.Nil(t, filter.EndDate)
			return []AssetOutput{
				{
					ID:          "asset-1",
					WalletID:    filter.WalletID,
					Type:        TypeStocks,
					Date:        "2026-07-15T12:00:00Z",
					Description: "Asset 1",
					Amount:      100000,
				},
				{
					ID:          "asset-2",
					WalletID:    filter.WalletID,
					Type:        TypeFIIs,
					Date:        "2026-07-20T12:00:00Z",
					Description: "Asset 2",
					Amount:      50000,
				},
			}, nil
		}
		handler := NewHandler(mockSvc)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, &mockWalletServiceForRoutes{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/assets", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[[]AssetOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 2)
		assert.Equal(t, "asset-1", response.Data[0].ID)
	})

	t.Run("success - parses type filter from query string", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).listAssetsFunc = func(ctx context.Context, filter AssetFilter) ([]AssetOutput, error) {
			assert.Equal(t, "wallet-id", filter.WalletID)
			assert.Equal(t, TypeStocks, filter.Type)
			return []AssetOutput{
				{
					ID:          "asset-1",
					WalletID:    filter.WalletID,
					Type:        TypeStocks,
					Date:        "2026-07-15T12:00:00Z",
					Description: "Asset 1",
					Amount:      100000,
				},
			}, nil
		}
		handler := NewHandler(mockSvc)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, &mockWalletServiceForRoutes{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/assets?type=stocks", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("success - parses start_date and end_date filters from query string", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).listAssetsFunc = func(ctx context.Context, filter AssetFilter) ([]AssetOutput, error) {
			assert.Equal(t, "wallet-id", filter.WalletID)
			require := filter.StartDate != nil && filter.EndDate != nil
			assert.True(t, require, "start_date and end_date must be parsed")
			assert.Equal(t, 2026, filter.StartDate.Year())
			assert.Equal(t, time.July, filter.StartDate.Month())
			assert.Equal(t, 1, filter.StartDate.Day())
			assert.Equal(t, 2026, filter.EndDate.Year())
			assert.Equal(t, time.July, filter.EndDate.Month())
			assert.Equal(t, 31, filter.EndDate.Day())
			return []AssetOutput{}, nil
		}
		handler := NewHandler(mockSvc)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, &mockWalletServiceForRoutes{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/assets?start_date=2026-07-01&end_date=2026-07-31", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error - returns 422 for invalid start_date format", func(t *testing.T) {
		mockSvc := newMockService(nil)
		handler := NewHandler(mockSvc)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, &mockWalletServiceForRoutes{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/assets?start_date=invalid", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 422 for invalid end_date format", func(t *testing.T) {
		mockSvc := newMockService(nil)
		handler := NewHandler(mockSvc)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, &mockWalletServiceForRoutes{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/assets?end_date=invalid", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 422 when start_date is after end_date", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).listAssetsFunc = func(ctx context.Context, filter AssetFilter) ([]AssetOutput, error) {
			return nil, ErrInvalidDateRange
		}
		handler := NewHandler(mockSvc)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, &mockWalletServiceForRoutes{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/assets?start_date=2026-08-01&end_date=2026-07-01", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 500 for internal server error", func(t *testing.T) {
		mockSvc := newMockService(nil)
		mockSvc.(*mockService).listAssetsFunc = func(ctx context.Context, filter AssetFilter) ([]AssetOutput, error) {
			return nil, errors.New("database error")
		}
		handler := NewHandler(mockSvc)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, &mockWalletServiceForRoutes{})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/assets", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	})
}
