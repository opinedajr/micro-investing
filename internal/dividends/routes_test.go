package dividends

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

	t.Run("success - registers dividends routes under /api/v1/wallets/:id/dividends", func(t *testing.T) {
		mockSvc := newMockService(serviceFuncs{
			createFunc: func(ctx context.Context, input CreateDividendInput) (*DividendOutput, error) {
				return &DividendOutput{ID: "dividend-id", Year: input.Year, Amount: input.Amount}, nil
			},
		})

		handler := NewHandler(mockSvc)
		walletService := &mockWalletServiceForRoutes{}

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, walletService)

		jsonBody, _ := json.Marshal(map[string]interface{}{
			"year":   2026,
			"amount": 150000,
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/dividends", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response api.Response[*DividendOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "dividend-id", response.Data.ID)
	})
}
