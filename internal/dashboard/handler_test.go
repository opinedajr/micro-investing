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
	summaryFunc    func(ctx context.Context, walletID string) (*SummaryOutput, error)
	allocationFunc func(ctx context.Context, walletID string) (*AllocationOutput, error)
	evolutionFunc  func(ctx context.Context, walletID string, input EvolutionInput) (*EvolutionOutput, error)
}

func (m *mockDashboardService) Summary(ctx context.Context, walletID string) (*SummaryOutput, error) {
	if m.summaryFunc != nil {
		return m.summaryFunc(ctx, walletID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDashboardService) Allocation(ctx context.Context, walletID string) (*AllocationOutput, error) {
	if m.allocationFunc != nil {
		return m.allocationFunc(ctx, walletID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDashboardService) Evolution(ctx context.Context, walletID string, input EvolutionInput) (*EvolutionOutput, error) {
	if m.evolutionFunc != nil {
		return m.evolutionFunc(ctx, walletID, input)
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

func TestHandler_Evolution(t *testing.T) {
	t.Run("success - returns evolution with query params", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		service := &mockDashboardService{
			evolutionFunc: func(ctx context.Context, walletID string, input EvolutionInput) (*EvolutionOutput, error) {
				assert.Equal(t, 2026, input.Year)
				assert.Equal(t, 2, input.Quarter)
				return &EvolutionOutput{
					Total: []EvolutionMonthOutput{
						{Year: 2026, Month: 4, Amount: 1200000},
						{Year: 2026, Month: 5, Amount: 1350000},
						{Year: 2026, Month: 6, Amount: 1500000},
					},
					ByCategory: EvolutionCategoryOutput{
						FixedIncome: []EvolutionMonthOutput{
							{Year: 2026, Month: 4, Amount: 500000},
							{Year: 2026, Month: 5, Amount: 500000},
							{Year: 2026, Month: 6, Amount: 550000},
						},
						Stocks: []EvolutionMonthOutput{
							{Year: 2026, Month: 4, Amount: 400000},
							{Year: 2026, Month: 5, Amount: 500000},
							{Year: 2026, Month: 6, Amount: 600000},
						},
						EmergencyReserve: []EvolutionMonthOutput{
							{Year: 2026, Month: 4, Amount: 300000},
							{Year: 2026, Month: 5, Amount: 350000},
							{Year: 2026, Month: 6, Amount: 350000},
						},
					},
				}, nil
			},
		}
		handler := NewHandler(service)

		r := gin.New()
		r.GET("/api/v1/wallets/:id/dashboard/evolution", handler.Evolution)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/evolution?year=2026&quarter=2", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[*EvolutionOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data.Total, 3)
		assert.Equal(t, int64(1500000), response.Data.Total[2].Amount)
		assert.Len(t, response.Data.ByCategory.FixedIncome, 3)
		assert.Equal(t, int64(550000), response.Data.ByCategory.FixedIncome[2].Amount)
	})

	t.Run("error - returns bad request when year is invalid", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		handler := NewHandler(&mockDashboardService{})

		r := gin.New()
		r.GET("/api/v1/wallets/:id/dashboard/evolution", handler.Evolution)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/evolution?year=abc", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns bad request when quarter is invalid string", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		handler := NewHandler(&mockDashboardService{})

		r := gin.New()
		r.GET("/api/v1/wallets/:id/dashboard/evolution", handler.Evolution)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/evolution?year=2026&quarter=abc", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns bad request when quarter is out of range", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		handler := NewHandler(&mockDashboardService{})

		r := gin.New()
		r.GET("/api/v1/wallets/:id/dashboard/evolution", handler.Evolution)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/evolution?year=2026&quarter=5", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns bad request when quarter is zero", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		handler := NewHandler(&mockDashboardService{})

		r := gin.New()
		r.GET("/api/v1/wallets/:id/dashboard/evolution", handler.Evolution)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/evolution?year=2026&quarter=0", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns bad request when quarter is negative", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		handler := NewHandler(&mockDashboardService{})

		r := gin.New()
		r.GET("/api/v1/wallets/:id/dashboard/evolution", handler.Evolution)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/evolution?year=2026&quarter=-1", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns bad request when quarter without year", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		handler := NewHandler(&mockDashboardService{})

		r := gin.New()
		r.GET("/api/v1/wallets/:id/dashboard/evolution", handler.Evolution)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/evolution?quarter=2", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns bad request when year is zero", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		handler := NewHandler(&mockDashboardService{})

		r := gin.New()
		r.GET("/api/v1/wallets/:id/dashboard/evolution", handler.Evolution)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/evolution?year=0&quarter=2", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns bad request when year is negative", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		handler := NewHandler(&mockDashboardService{})

		r := gin.New()
		r.GET("/api/v1/wallets/:id/dashboard/evolution", handler.Evolution)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/evolution?year=-2026&quarter=2", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns internal server error when service fails", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		service := &mockDashboardService{
			evolutionFunc: func(ctx context.Context, walletID string, input EvolutionInput) (*EvolutionOutput, error) {
				return nil, errors.New("service error")
			},
		}
		handler := NewHandler(service)

		r := gin.New()
		r.GET("/api/v1/wallets/:id/dashboard/evolution", handler.Evolution)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/evolution", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	})
}

func TestHandler_Allocation(t *testing.T) {
	t.Run("success - returns allocation", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		service := &mockDashboardService{
			allocationFunc: func(ctx context.Context, walletID string) (*AllocationOutput, error) {
				return &AllocationOutput{
					Items: []AllocationItem{
						{Type: "stocks", Amount: 500000, Percentage: 33.33},
						{Type: "fixed_income", Amount: 500000, Percentage: 33.33},
						{Type: "emergency_reserve", Amount: 250000, Percentage: 16.67},
						{Type: "liquid_cash", Amount: 250000, Percentage: 16.67},
					},
					Total: 1500000,
				}, nil
			},
		}
		handler := NewHandler(service)

		r := gin.New()
		r.GET("/api/v1/wallets/:id/dashboard/allocation", handler.Allocation)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/allocation", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[*AllocationOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data.Items, 4)
		assert.Equal(t, int64(1500000), response.Data.Total)
		assert.Equal(t, "stocks", response.Data.Items[0].Type)
		assert.Equal(t, int64(500000), response.Data.Items[0].Amount)
		assert.InDelta(t, 33.33, response.Data.Items[0].Percentage, 0.01)
	})

	t.Run("error - returns internal server error when service fails", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		service := &mockDashboardService{
			allocationFunc: func(ctx context.Context, walletID string) (*AllocationOutput, error) {
				return nil, errors.New("service error")
			},
		}
		handler := NewHandler(service)

		r := gin.New()
		r.GET("/api/v1/wallets/:id/dashboard/allocation", handler.Allocation)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dashboard/allocation", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	})
}
