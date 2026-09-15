package dashboard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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

	t.Run("success - registers dashboard summary route under /api/v1/wallets/:id/dashboard/summary", func(t *testing.T) {
		service := &mockDashboardService{
			summaryFunc: func(ctx context.Context, walletID string) (*SummaryOutput, error) {
				return &SummaryOutput{
					CurrentPatrimony: 1500000,
					YearlyDividends:  0,
					StocksInvested:   500000,
				}, nil
			},
		}
		handler := NewHandler(service)
		walletService := &mockWalletServiceForRoutes{}

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, walletService)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/summary", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("success - registers dashboard allocation route under /api/v1/wallets/:id/dashboard/allocation", func(t *testing.T) {
		service := &mockDashboardService{
			allocationFunc: func(ctx context.Context, walletID string) (*AllocationOutput, error) {
				return &AllocationOutput{
					Items: []AllocationItem{
						{Type: "stocks", Amount: 500000, Percentage: 100},
					},
					Total: 500000,
				}, nil
			},
		}
		handler := NewHandler(service)
		walletService := &mockWalletServiceForRoutes{}

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler, walletService)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/allocation", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
