package dashboard

import (
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

type mockDashboardService struct {
	summaryFunc func(ctx context.Context, walletID string) (*SummaryOutput, error)
}

func (m *mockDashboardService) Summary(ctx context.Context, walletID string) (*SummaryOutput, error) {
	if m.summaryFunc != nil {
		return m.summaryFunc(ctx, walletID)
	}
	return nil, errors.New("not implemented")
}

func TestHandler_Summary(t *testing.T) {
	t.Run("success - returns summary", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
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

		r := gin.New()
		r.GET("/api/v1/wallets/:id/dashboard/summary", handler.Summary)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/summary", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[*SummaryOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, int64(1500000), response.Data.CurrentPatrimony)
		assert.Equal(t, int64(0), response.Data.YearlyDividends)
		assert.Equal(t, int64(500000), response.Data.StocksInvested)
	})

	t.Run("error - returns internal server error when service fails", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		service := &mockDashboardService{
			summaryFunc: func(ctx context.Context, walletID string) (*SummaryOutput, error) {
				return nil, errors.New("service error")
			},
		}
		handler := NewHandler(service)

		r := gin.New()
		r.GET("/api/v1/wallets/:id/dashboard/summary", handler.Summary)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/summary", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	})
}
