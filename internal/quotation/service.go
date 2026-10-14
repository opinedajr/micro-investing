package quotation

import (
	"context"
	"errors"
	"fmt"

	"github.com/opinedajr/micro-investing/internal/shared/logger"
	"github.com/opinedajr/micro-investing/internal/stock"
)

type Service interface {
	SyncCurrentPrices(ctx context.Context, input SyncInput) (*SyncOutput, error)
}

type quotationService struct {
	provider  Provider
	repo      Repository
	stockRepo stock.Repository
	batchSize int
	logger    logger.Logger
}

func NewService(provider Provider, repo Repository, stockRepo stock.Repository, batchSize int, logger logger.Logger) Service {
	if batchSize < 1 {
		batchSize = 1
	}
	return &quotationService{
		provider:  provider,
		repo:      repo,
		stockRepo: stockRepo,
		batchSize: batchSize,
		logger:    logger,
	}
}

func (s *quotationService) SyncCurrentPrices(ctx context.Context, input SyncInput) (*SyncOutput, error) {
	tickers, tickerToStockID, err := s.resolveTickers(ctx, input.Tickers)
	if err != nil {
		return nil, err
	}

	var quotes []QuoteOutput
	var failed []string
	for _, batch := range chunkTickers(tickers, s.batchSize) {
		batchQuotes, err := s.provider.FetchQuotes(ctx, batch)
		if err != nil {
			s.logger.Warn(ctx, "quote fetch failed for batch", "tickers", batch, "error", err)
			failed = append(failed, batch...)
			continue
		}
		quotes = append(quotes, batchQuotes...)
	}

	skipped := skippedTickers(tickers, quotes, failed)
	if len(skipped) > 0 {
		s.logger.Warn(ctx, "tickers skipped by provider", "tickers", skipped)
	}

	prices := toCurrentPrices(quotes, tickerToStockID)

	if len(prices) == 0 && len(failed) > 0 {
		s.logger.Error(ctx, "current prices sync failed for every batch", "failed", failed)
		return nil, ErrQuotesFetchFailed
	}

	if err := s.repo.UpsertCurrentPrices(ctx, prices); err != nil {
		return nil, err
	}

	s.logger.Info(ctx, "current prices sync finished", "updated", len(prices), "skipped", skipped, "failed", failed)

	return &SyncOutput{Updated: len(prices), Skipped: skipped, Failed: failed}, nil
}

func (s *quotationService) resolveTickers(ctx context.Context, requested []string) ([]string, map[string]string, error) {
	if len(requested) > 0 {
		tickers := make([]string, 0, len(requested))
		tickerToStockID := make(map[string]string, len(requested))
		for _, ticker := range requested {
			found, err := s.stockRepo.FindByTicker(ctx, ticker)
			if err != nil {
				if errors.Is(err, stock.ErrStockNotFound) {
					return nil, nil, fmt.Errorf("%w: %s", ErrUnknownTicker, ticker)
				}
				return nil, nil, err
			}
			tickers = append(tickers, found.Ticker)
			tickerToStockID[found.Ticker] = found.ID
		}
		return tickers, tickerToStockID, nil
	}

	stocks, err := s.stockRepo.List(ctx)
	if err != nil {
		return nil, nil, err
	}

	tickers := make([]string, 0, len(stocks))
	tickerToStockID := make(map[string]string, len(stocks))
	for i := range stocks {
		tickers = append(tickers, stocks[i].Ticker)
		tickerToStockID[stocks[i].Ticker] = stocks[i].ID
	}
	return tickers, tickerToStockID, nil
}

func chunkTickers(tickers []string, size int) [][]string {
	if len(tickers) == 0 {
		return nil
	}
	batches := make([][]string, 0, (len(tickers)+size-1)/size)
	for start := 0; start < len(tickers); start += size {
		end := min(start+size, len(tickers))
		batches = append(batches, tickers[start:end])
	}
	return batches
}

func skippedTickers(requested []string, quotes []QuoteOutput, failed []string) []string {
	obtained := make(map[string]bool, len(quotes))
	for _, quote := range quotes {
		obtained[quote.Ticker] = true
	}
	fetchFailed := make(map[string]bool, len(failed))
	for _, ticker := range failed {
		fetchFailed[ticker] = true
	}
	skipped := make([]string, 0)
	for _, ticker := range requested {
		if !obtained[ticker] && !fetchFailed[ticker] {
			skipped = append(skipped, ticker)
		}
	}
	return skipped
}

func toCurrentPrices(quotes []QuoteOutput, tickerToStockID map[string]string) []CurrentPrice {
	prices := make([]CurrentPrice, 0, len(quotes))
	for _, quote := range quotes {
		stockID, ok := tickerToStockID[quote.Ticker]
		if !ok {
			continue
		}
		prices = append(prices, CurrentPrice{StockID: stockID, Price: quote.Price})
	}
	return prices
}
