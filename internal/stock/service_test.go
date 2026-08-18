package stock

import (
	"context"
	"errors"

	"github.com/stretchr/testify/assert"
	"testing"
)

type mockRepository struct {
	listFunc          func(ctx context.Context) ([]Stock, error)
	findByTickerFunc  func(ctx context.Context, ticker string) (*Stock, error)
	createFunc        func(ctx context.Context, stock *Stock) error
	seedFunc          func(ctx context.Context, stocks []Stock, force bool) (int, int, int, error)
}

func (m *mockRepository) List(ctx context.Context) ([]Stock, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, nil
}

func (m *mockRepository) FindByTicker(ctx context.Context, ticker string) (*Stock, error) {
	if m.findByTickerFunc != nil {
		return m.findByTickerFunc(ctx, ticker)
	}
	return nil, errors.New("not found")
}

func (m *mockRepository) Create(ctx context.Context, stock *Stock) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, stock)
	}
	return nil
}

func (m *mockRepository) Seed(ctx context.Context, stocks []Stock, force bool) (int, int, int, error) {
	if m.seedFunc != nil {
		return m.seedFunc(ctx, stocks, force)
	}
	return 0, 0, 0, nil
}

func TestStockService_ListStocks_Success(t *testing.T) {
	mockRepo := &mockRepository{
		listFunc: func(ctx context.Context) ([]Stock, error) {
			return []Stock{
				{ID: "1", Ticker: "PETR4", Name: "Petrobras", Sector: "Energia", Rank: 10},
				{ID: "2", Ticker: "VALE3", Name: "Vale", Sector: "Mineração", Rank: 10},
			}, nil
		},
	}

	service := NewStockService(mockRepo)
	stocks, err := service.ListStocks(context.Background())

	assert.NoError(t, err)
	assert.Len(t, stocks, 2)
	assert.Equal(t, "PETR4", stocks[0].Ticker)
	assert.Equal(t, "VALE3", stocks[1].Ticker)
}

func TestStockService_ListStocks_Error(t *testing.T) {
	mockRepo := &mockRepository{
		listFunc: func(ctx context.Context) ([]Stock, error) {
			return nil, errors.New("database error")
		},
	}

	service := NewStockService(mockRepo)
	stocks, err := service.ListStocks(context.Background())

	assert.Nil(t, stocks)
	assert.Error(t, err)
	assert.Equal(t, ErrFailedToList, err)
}

func TestStockService_GetStockByTicker_Success(t *testing.T) {
	mockRepo := &mockRepository{
		findByTickerFunc: func(ctx context.Context, ticker string) (*Stock, error) {
			return &Stock{ID: "1", Ticker: "PETR4", Name: "Petrobras", Sector: "Energia", Rank: 10}, nil
		},
	}

	service := NewStockService(mockRepo)
	stock, err := service.GetStockByTicker(context.Background(), "PETR4")

	assert.NoError(t, err)
	assert.NotNil(t, stock)
	assert.Equal(t, "PETR4", stock.Ticker)
	assert.Equal(t, "Petrobras", stock.Name)
}

func TestStockService_GetStockByTicker_NotFound(t *testing.T) {
	mockRepo := &mockRepository{
		findByTickerFunc: func(ctx context.Context, ticker string) (*Stock, error) {
			return nil, ErrStockNotFound
		},
	}

	service := NewStockService(mockRepo)
	stock, err := service.GetStockByTicker(context.Background(), "XXXX")

	assert.Nil(t, stock)
	assert.Error(t, err)
	assert.Equal(t, ErrStockNotFound, err)
}

func TestStockService_GetStockByTicker_Error(t *testing.T) {
	mockRepo := &mockRepository{
		findByTickerFunc: func(ctx context.Context, ticker string) (*Stock, error) {
			return nil, errors.New("database error")
		},
	}

	service := NewStockService(mockRepo)
	stock, err := service.GetStockByTicker(context.Background(), "PETR4")

	assert.Nil(t, stock)
	assert.Error(t, err)
	assert.Equal(t, ErrFailedToFind, err)
}

func TestStockService_SeedStocks_Success(t *testing.T) {
	mockRepo := &mockRepository{
		seedFunc: func(ctx context.Context, stocks []Stock, force bool) (int, int, int, error) {
			return 20, 0, 0, nil
		},
	}

	service := NewStockService(mockRepo)
	inserted, updated, skipped, err := service.SeedStocks(context.Background(), false)

	assert.NoError(t, err)
	assert.Equal(t, 20, inserted)
	assert.Equal(t, 0, updated)
	assert.Equal(t, 0, skipped)
}

func TestStockService_SeedStocks_Force(t *testing.T) {
	mockRepo := &mockRepository{
		seedFunc: func(ctx context.Context, stocks []Stock, force bool) (int, int, int, error) {
			return 0, 15, 5, nil
		},
	}

	service := NewStockService(mockRepo)
	inserted, updated, skipped, err := service.SeedStocks(context.Background(), true)

	assert.NoError(t, err)
	assert.Equal(t, 0, inserted)
	assert.Equal(t, 15, updated)
	assert.Equal(t, 5, skipped)
}
