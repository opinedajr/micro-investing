package position

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/api"
	"github.com/opinedajr/micro-investing/internal/wallet"
	"github.com/stretchr/testify/assert"
)

type mockWalletServiceForRoutes struct{}

func (m *mockWalletServiceForRoutes) Create(ctx context.Context, input wallet.CreateWalletInput) (*wallet.WalletOutput, error) {
	return nil, nil
}

func (m *mockWalletServiceForRoutes) List(ctx context.Context, userID string) ([]wallet.WalletOutput, error) {
	return nil, nil
}

func (m *mockWalletServiceForRoutes) Find(ctx context.Context, id string) (*wallet.WalletOutput, error) {
	return &wallet.WalletOutput{ID: id}, nil
}

func (m *mockWalletServiceForRoutes) Update(ctx context.Context, id string, input wallet.UpdateWalletInput) (*wallet.WalletOutput, error) {
	return nil, nil
}

func (m *mockWalletServiceForRoutes) Delete(ctx context.Context, id string) error {
	return nil
}

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - registers position list route under /api/v1/wallets/:id/positions", func(t *testing.T) {
		mockSvc := &mockService{
			listFunc: func(ctx context.Context, filter PositionFilter) ([]PositionOutput, error) {
				return []PositionOutput{{ID: "p1", WalletID: filter.WalletID, StockID: "s1"}}, nil
			},
		}

		handler := NewHandler(mockSvc)
		walletService := &mockWalletServiceForRoutes{}

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, walletService)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/positions", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[[]PositionOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, "p1", response.Data[0].ID)
	})

	t.Run("success - registers position find route under /api/v1/wallets/:id/positions/:positionId", func(t *testing.T) {
		mockSvc := &mockService{
			findFunc: func(ctx context.Context, walletID string, id string) (*PositionOutput, error) {
				return &PositionOutput{ID: id, WalletID: walletID, StockID: "s1"}, nil
			},
		}

		handler := NewHandler(mockSvc)
		walletService := &mockWalletServiceForRoutes{}

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, walletService)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/positions/p1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[*PositionOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "p1", response.Data.ID)
	})

	t.Run("success - registers position create route under /api/v1/wallets/:id/positions", func(t *testing.T) {
		mockSvc := newMockServiceWithCreate(func(ctx context.Context, input CreatePositionInput) (*PositionOutput, error) {
			return &PositionOutput{
				ID:           "position-id",
				WalletID:     input.WalletID,
				StockID:      input.StockID,
				Quantity:     input.Quantity,
				AveragePrice: input.AveragePrice,
				CreatedAt:    "2026-08-26T12:00:00Z",
				UpdatedAt:    "2026-08-26T12:00:00Z",
			}, nil
		})

		handler := NewHandler(mockSvc)
		walletService := &mockWalletServiceForRoutes{}

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, walletService)

		body := map[string]interface{}{
			"stock_id":      "stock-id",
			"quantity":      10,
			"average_price": 1000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/positions", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response api.Response[*PositionOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "position-id", response.Data.ID)
	})

	t.Run("success - registers position update route under /api/v1/wallets/:id/positions/:positionId", func(t *testing.T) {
		mockSvc := &mockService{
			updateFunc: func(ctx context.Context, input UpdatePositionInput) (*PositionOutput, error) {
				return &PositionOutput{
					ID:           input.PositionID,
					WalletID:     input.WalletID,
					StockID:      "s1",
					Quantity:     input.Quantity,
					AveragePrice: input.AveragePrice,
					UpdatedAt:    "2026-08-26T12:00:00Z",
				}, nil
			},
		}

		handler := NewHandler(mockSvc)
		walletService := &mockWalletServiceForRoutes{}

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, walletService)

		body := map[string]interface{}{
			"quantity":      20,
			"average_price": 2000,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/v1/wallets/wallet-id/positions/p1", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[*PositionOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "p1", response.Data.ID)
	})
}
