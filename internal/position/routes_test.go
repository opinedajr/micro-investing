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
}
