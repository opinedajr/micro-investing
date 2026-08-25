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
	findByTickerFunc func(ctx context.Context, ticker string) (*StockOutput, error)
	listFunc         func(ctx context.Context) ([]StockOutput, error)
	seedFunc         func(ctx context.Context, input SeedInput) error
}

func (m *mockService) FindByTicker(ctx context.Context, ticker string) (*StockOutput, error) {
	if m.findByTickerFunc != nil {
		return m.findByTickerFunc(ctx, ticker)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) List(ctx context.Context) ([]StockOutput, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) Seed(ctx context.Context, input SeedInput) error {
	if m.seedFunc != nil {
		return m.seedFunc(ctx, input)
	}
	return errors.New("not implemented")
}

func newMockServiceWithList(listFunc func(ctx context.Context) ([]StockOutput, error)) Service {
	return &mockService{listFunc: listFunc}
}

func newMockServiceWithFind(findByTickerFunc func(ctx context.Context, ticker string) (*StockOutput, error)) Service {
	return &mockService{findByTickerFunc: findByTickerFunc}
}

func TestHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - returns stock list", func(t *testing.T) {
		service := newMockServiceWithList(func(ctx context.Context) ([]StockOutput, error) {
			return []StockOutput{
				{ID: "1", Ticker: "PETR4", Name: "Petrobras PN", Sector: "Petróleo", Rank: 10},
				{ID: "2", Ticker: "VALE3", Name: "Vale ON", Sector: "Mineração", Rank: 10},
			}, nil
		})

		handler := NewHandler(service)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/stocks", nil)

		handler.List(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[[]StockOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 2)
		assert.Equal(t, "PETR4", response.Data[0].Ticker)
	})

	t.Run("error - returns 500 on service error", func(t *testing.T) {
		service := newMockServiceWithList(func(ctx context.Context) ([]StockOutput, error) {
			return nil, errors.New("database error")
		})

		handler := NewHandler(service)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/stocks", nil)

		handler.List(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	})
}

func TestHandler_FindByTicker(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - returns stock detail", func(t *testing.T) {
		service := newMockServiceWithFind(func(ctx context.Context, ticker string) (*StockOutput, error) {
			return &StockOutput{
				ID:     "stock-id",
				Ticker: ticker,
				Name:   "Petrobras PN",
				Sector: "Petróleo",
				Rank:   10,
			}, nil
		})

		handler := NewHandler(service)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/stocks/PETR4", nil)
		c.Params = []gin.Param{{Key: "ticker", Value: "PETR4"}}

		handler.FindByTicker(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[*StockOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "PETR4", response.Data.Ticker)
		assert.Equal(t, "Petrobras PN", response.Data.Name)
	})

	t.Run("error - returns 404 when stock not found", func(t *testing.T) {
		service := newMockServiceWithFind(func(ctx context.Context, ticker string) (*StockOutput, error) {
			return nil, ErrStockNotFound
		})

		handler := NewHandler(service)
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/stocks/INVALID", nil)
		c.Params = []gin.Param{{Key: "ticker", Value: "INVALID"}}

		handler.FindByTicker(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "STOCK_NOT_FOUND", response.Error.Code)
	})
}
