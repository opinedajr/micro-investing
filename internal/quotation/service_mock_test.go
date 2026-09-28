package quotation

import (
	"context"
	"errors"

	"github.com/opinedajr/micro-investing/internal/stock"
)

type mockProvider struct {
	fetchQuotesFunc func(ctx context.Context, tickers []string) ([]QuoteOutput, error)
}

func (m *mockProvider) FetchQuotes(ctx context.Context, tickers []string) ([]QuoteOutput, error) {
	if m.fetchQuotesFunc != nil {
		return m.fetchQuotesFunc(ctx, tickers)
	}
	return nil, nil
}

type mockRepository struct {
	upsertCurrentPricesFunc func(ctx context.Context, prices []CurrentPrice) error
}

func (m *mockRepository) UpsertCurrentPrices(ctx context.Context, prices []CurrentPrice) error {
	if m.upsertCurrentPricesFunc != nil {
		return m.upsertCurrentPricesFunc(ctx, prices)
	}
	return nil
}

type mockStockRepository struct {
	listFunc         func(ctx context.Context) ([]stock.Stock, error)
	findByTickerFunc func(ctx context.Context, ticker string) (*stock.Stock, error)
}

func (m *mockStockRepository) Create(ctx context.Context, s *stock.Stock) error {
	return errors.New("not implemented")
}

func (m *mockStockRepository) FindByTicker(ctx context.Context, ticker string) (*stock.Stock, error) {
	if m.findByTickerFunc != nil {
		return m.findByTickerFunc(ctx, ticker)
	}
	return nil, stock.ErrStockNotFound
}

func (m *mockStockRepository) FindByID(ctx context.Context, id string) (*stock.Stock, error) {
	return nil, errors.New("not implemented")
}

func (m *mockStockRepository) FindByIDs(ctx context.Context, ids []string) ([]stock.Stock, error) {
	return nil, errors.New("not implemented")
}

func (m *mockStockRepository) List(ctx context.Context) ([]stock.Stock, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, nil
}

func (m *mockStockRepository) Seed(ctx context.Context, stocks []stock.Stock, force bool) error {
	return errors.New("not implemented")
}
