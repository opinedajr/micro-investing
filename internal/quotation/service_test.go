package quotation

import (
	"context"
	"errors"
	"testing"

	"github.com/opinedajr/micro-investing/internal/shared/logger"
	"github.com/opinedajr/micro-investing/internal/stock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func syncTestStocks() []stock.Stock {
	return []stock.Stock{
		{ID: "id-petr4", Ticker: "PETR4"},
		{ID: "id-vale3", Ticker: "VALE3"},
		{ID: "id-itub4", Ticker: "ITUB4"},
		{ID: "id-bbas3", Ticker: "BBAS3"},
	}
}

func stockFinderByTicker(stocks []stock.Stock) func(ctx context.Context, ticker string) (*stock.Stock, error) {
	return func(ctx context.Context, ticker string) (*stock.Stock, error) {
		for i := range stocks {
			if stocks[i].Ticker == ticker {
				return &stocks[i], nil
			}
		}
		return nil, stock.ErrStockNotFound
	}
}

func quotesFor(tickers []string, price int64) []QuoteOutput {
	quotes := make([]QuoteOutput, 0, len(tickers))
	for _, ticker := range tickers {
		quotes = append(quotes, QuoteOutput{Ticker: ticker, Price: price})
	}
	return quotes
}

func newSyncTestService(provider Provider, repo Repository, stockRepo stock.Repository, batchSize int) Service {
	return NewService(provider, repo, stockRepo, batchSize, logger.NewLogger("error"))
}

func TestSyncCurrentPricesSuccess(t *testing.T) {
	stocks := syncTestStocks()[:2]
	stockRepo := &mockStockRepository{findByTickerFunc: stockFinderByTicker(stocks)}
	provider := &mockProvider{
		fetchQuotesFunc: func(ctx context.Context, tickers []string) ([]QuoteOutput, error) {
			return quotesFor(tickers, 4118), nil
		},
	}
	var upserted [][]CurrentPrice
	repo := &mockRepository{
		upsertCurrentPricesFunc: func(ctx context.Context, prices []CurrentPrice) error {
			upserted = append(upserted, prices)
			return nil
		},
	}

	service := newSyncTestService(provider, repo, stockRepo, 2)
	output, err := service.SyncCurrentPrices(context.Background(), SyncInput{Tickers: []string{"PETR4", "VALE3"}})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 2, output.Updated)
	assert.Empty(t, output.Skipped)
	assert.Empty(t, output.Failed)
	require.Len(t, upserted, 1)
	assert.ElementsMatch(t, []CurrentPrice{
		{StockID: "id-petr4", Price: 4118},
		{StockID: "id-vale3", Price: 4118},
	}, upserted[0])
}

func TestSyncCurrentPricesPartialSkip(t *testing.T) {
	stocks := syncTestStocks()[:3]
	stockRepo := &mockStockRepository{findByTickerFunc: stockFinderByTicker(stocks)}
	provider := &mockProvider{
		fetchQuotesFunc: func(ctx context.Context, tickers []string) ([]QuoteOutput, error) {
			return quotesFor([]string{"PETR4", "VALE3"}, 2500), nil
		},
	}
	var upserted []CurrentPrice
	repo := &mockRepository{
		upsertCurrentPricesFunc: func(ctx context.Context, prices []CurrentPrice) error {
			upserted = prices
			return nil
		},
	}

	service := newSyncTestService(provider, repo, stockRepo, 3)
	output, err := service.SyncCurrentPrices(context.Background(), SyncInput{Tickers: []string{"PETR4", "VALE3", "ITUB4"}})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 2, output.Updated)
	assert.Equal(t, []string{"ITUB4"}, output.Skipped)
	assert.Empty(t, output.Failed)
	require.Len(t, upserted, 2)
	for _, price := range upserted {
		assert.NotEqual(t, "id-itub4", price.StockID)
	}
}

func TestSyncCurrentPricesBatchFailureContinues(t *testing.T) {
	stocks := syncTestStocks()
	stockRepo := &mockStockRepository{findByTickerFunc: stockFinderByTicker(stocks)}
	var fetchCalls [][]string
	provider := &mockProvider{
		fetchQuotesFunc: func(ctx context.Context, tickers []string) ([]QuoteOutput, error) {
			fetchCalls = append(fetchCalls, tickers)
			if tickers[0] == "PETR4" {
				return nil, ErrQuotesFetchFailed
			}
			return quotesFor(tickers, 3000), nil
		},
	}
	var upserted []CurrentPrice
	repo := &mockRepository{
		upsertCurrentPricesFunc: func(ctx context.Context, prices []CurrentPrice) error {
			upserted = prices
			return nil
		},
	}

	service := newSyncTestService(provider, repo, stockRepo, 2)
	output, err := service.SyncCurrentPrices(context.Background(), SyncInput{Tickers: []string{"PETR4", "VALE3", "ITUB4", "BBAS3"}})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Len(t, fetchCalls, 2)
	assert.Equal(t, []string{"PETR4", "VALE3"}, output.Failed)
	assert.Equal(t, 2, output.Updated)
	assert.Empty(t, output.Skipped)
	require.Len(t, upserted, 2)
	assert.ElementsMatch(t, []CurrentPrice{
		{StockID: "id-itub4", Price: 3000},
		{StockID: "id-bbas3", Price: 3000},
	}, upserted)
}

func TestSyncCurrentPricesTotalFailure(t *testing.T) {
	stocks := syncTestStocks()
	stockRepo := &mockStockRepository{findByTickerFunc: stockFinderByTicker(stocks)}
	provider := &mockProvider{
		fetchQuotesFunc: func(ctx context.Context, tickers []string) ([]QuoteOutput, error) {
			return nil, ErrQuotesFetchFailed
		},
	}
	upsertCalls := 0
	repo := &mockRepository{
		upsertCurrentPricesFunc: func(ctx context.Context, prices []CurrentPrice) error {
			upsertCalls++
			return nil
		},
	}

	service := newSyncTestService(provider, repo, stockRepo, 2)
	output, err := service.SyncCurrentPrices(context.Background(), SyncInput{Tickers: []string{"PETR4", "VALE3", "ITUB4", "BBAS3"}})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrQuotesFetchFailed)
	assert.Nil(t, output)
	assert.Equal(t, 0, upsertCalls)
}

func TestSyncCurrentPricesUnknownTicker(t *testing.T) {
	stocks := syncTestStocks()[:1]
	stockRepo := &mockStockRepository{findByTickerFunc: stockFinderByTicker(stocks)}
	fetchCalls := 0
	provider := &mockProvider{
		fetchQuotesFunc: func(ctx context.Context, tickers []string) ([]QuoteOutput, error) {
			fetchCalls++
			return quotesFor(tickers, 1000), nil
		},
	}
	upsertCalls := 0
	repo := &mockRepository{
		upsertCurrentPricesFunc: func(ctx context.Context, prices []CurrentPrice) error {
			upsertCalls++
			return nil
		},
	}

	service := newSyncTestService(provider, repo, stockRepo, 2)
	output, err := service.SyncCurrentPrices(context.Background(), SyncInput{Tickers: []string{"PETR4", "FAKE4"}})

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrUnknownTicker)
	assert.Nil(t, output)
	assert.Equal(t, 0, fetchCalls)
	assert.Equal(t, 0, upsertCalls)
}

func TestSyncCurrentPricesEmptyInputListsAllStocks(t *testing.T) {
	stocks := syncTestStocks()
	stockRepo := &mockStockRepository{listFunc: func(ctx context.Context) ([]stock.Stock, error) {
		return stocks, nil
	}}
	var fetchCalls [][]string
	provider := &mockProvider{
		fetchQuotesFunc: func(ctx context.Context, tickers []string) ([]QuoteOutput, error) {
			fetchCalls = append(fetchCalls, tickers)
			return quotesFor(tickers, 5000), nil
		},
	}
	var upserted []CurrentPrice
	repo := &mockRepository{
		upsertCurrentPricesFunc: func(ctx context.Context, prices []CurrentPrice) error {
			upserted = prices
			return nil
		},
	}

	service := newSyncTestService(provider, repo, stockRepo, 4)
	output, err := service.SyncCurrentPrices(context.Background(), SyncInput{})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 4, output.Updated)
	assert.Empty(t, output.Skipped)
	assert.Empty(t, output.Failed)
	require.Len(t, fetchCalls, 1)
	assert.ElementsMatch(t, []string{"PETR4", "VALE3", "ITUB4", "BBAS3"}, fetchCalls[0])
	require.Len(t, upserted, 4)
}

func TestSyncCurrentPricesSingleUpsertTransaction(t *testing.T) {
	stocks := syncTestStocks()[:3]
	stockRepo := &mockStockRepository{findByTickerFunc: stockFinderByTicker(stocks)}
	fetchCalls := 0
	provider := &mockProvider{
		fetchQuotesFunc: func(ctx context.Context, tickers []string) ([]QuoteOutput, error) {
			fetchCalls++
			return quotesFor(tickers, 7000), nil
		},
	}
	var upsertCalls [][]CurrentPrice
	repo := &mockRepository{
		upsertCurrentPricesFunc: func(ctx context.Context, prices []CurrentPrice) error {
			upsertCalls = append(upsertCalls, prices)
			return nil
		},
	}

	service := newSyncTestService(provider, repo, stockRepo, 1)
	output, err := service.SyncCurrentPrices(context.Background(), SyncInput{Tickers: []string{"PETR4", "VALE3", "ITUB4"}})

	require.NoError(t, err)
	require.NotNil(t, output)
	assert.Equal(t, 3, fetchCalls)
	require.Len(t, upsertCalls, 1)
	require.Len(t, upsertCalls[0], 3)
	assert.Equal(t, []CurrentPrice{
		{StockID: "id-petr4", Price: 7000},
		{StockID: "id-vale3", Price: 7000},
		{StockID: "id-itub4", Price: 7000},
	}, upsertCalls[0])
}

func TestSyncCurrentPricesUpsertError(t *testing.T) {
	stocks := syncTestStocks()[:1]
	stockRepo := &mockStockRepository{findByTickerFunc: stockFinderByTicker(stocks)}
	provider := &mockProvider{
		fetchQuotesFunc: func(ctx context.Context, tickers []string) ([]QuoteOutput, error) {
			return quotesFor(tickers, 1000), nil
		},
	}
	upsertError := errors.New("upsert failed")
	repo := &mockRepository{
		upsertCurrentPricesFunc: func(ctx context.Context, prices []CurrentPrice) error {
			return upsertError
		},
	}

	service := newSyncTestService(provider, repo, stockRepo, 1)
	output, err := service.SyncCurrentPrices(context.Background(), SyncInput{Tickers: []string{"PETR4"}})

	require.Error(t, err)
	assert.ErrorIs(t, err, upsertError)
	assert.Nil(t, output)
}
