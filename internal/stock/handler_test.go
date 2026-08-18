package stock

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

type mockService struct {
	listFunc        func(ctx context.Context) ([]Stock, error)
	getByTickerFunc func(ctx context.Context, ticker string) (*Stock, error)
}

func (m *mockService) ListStocks(ctx context.Context) ([]Stock, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, nil
}

func (m *mockService) GetStockByTicker(ctx context.Context, ticker string) (*Stock, error) {
	if m.getByTickerFunc != nil {
		return m.getByTickerFunc(ctx, ticker)
	}
	return nil, nil
}

func (m *mockService) SeedStocks(ctx context.Context, force bool) (int, int, int, error) {
	return 0, 0, 0, nil
}

func TestHandler_ListStocks(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - lists all stocks", func(t *testing.T) {
		mockSvc := &mockService{
			listFunc: func(ctx context.Context) ([]Stock, error) {
				return []Stock{
					{ID: "1", Ticker: "PETR4", Name: "Petrobras", Sector: "Energia", Rank: 10},
					{ID: "2", Ticker: "VALE3", Name: "Vale", Sector: "Mineração", Rank: 10},
				}, nil
			},
		}

		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/stocks", nil)

		handler.ListStocks(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response ListOutput
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Stocks, 2)
		assert.Equal(t, "PETR4", response.Stocks[0].Ticker)
		assert.Equal(t, "VALE3", response.Stocks[1].Ticker)
	})

	t.Run("error - returns 500 on service error", func(t *testing.T) {
		mockSvc := &mockService{
			listFunc: func(ctx context.Context) ([]Stock, error) {
				return nil, errors.New("database error")
			},
		}

		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/stocks", nil)

		handler.ListStocks(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Error)
	})
}

func TestHandler_GetStockByTicker(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - gets stock by ticker", func(t *testing.T) {
		mockSvc := &mockService{
			getByTickerFunc: func(ctx context.Context, ticker string) (*Stock, error) {
				return &Stock{
					ID:     "1",
					Ticker: "PETR4",
					Name:   "Petrobras",
					Sector: "Energia",
					Rank:   10,
				}, nil
			},
		}

		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/stocks/PETR4", nil)
		c.Params = gin.Params{{Key: "ticker", Value: "PETR4"}}

		handler.GetStockByTicker(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response StockOutput
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "PETR4", response.Ticker)
		assert.Equal(t, "Petrobras", response.Name)
		assert.Equal(t, "Energia", response.Sector)
		assert.Equal(t, int8(10), response.Rank)
	})

	t.Run("error - returns 404 when stock not found", func(t *testing.T) {
		mockSvc := &mockService{
			getByTickerFunc: func(ctx context.Context, ticker string) (*Stock, error) {
				return nil, ErrStockNotFound
			},
		}

		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/stocks/XXXX", nil)
		c.Params = gin.Params{{Key: "ticker", Value: "XXXX"}}

		handler.GetStockByTicker(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "STOCK_NOT_FOUND", response.Error.Code)
	})

	t.Run("error - returns 500 on service error", func(t *testing.T) {
		mockSvc := &mockService{
			getByTickerFunc: func(ctx context.Context, ticker string) (*Stock, error) {
				return nil, errors.New("database error")
			},
		}

		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/stocks/PETR4", nil)
		c.Params = gin.Params{{Key: "ticker", Value: "PETR4"}}

		handler.GetStockByTicker(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Error)
	})
}
