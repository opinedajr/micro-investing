package brapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strings"

	"github.com/opinedajr/micro-investing/internal/quotation"
	"github.com/opinedajr/micro-investing/internal/shared/config"
)

type Client struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

type quoteResponse struct {
	Results []quoteResult `json:"results"`
}

type quoteResult struct {
	RequestedSymbol string     `json:"requestedSymbol"`
	Data            quotePrice `json:"data"`
}

type quotePrice struct {
	RegularMarketPrice float64 `json:"regularMarketPrice"`
}

func NewClient(cfg config.BrapiConfig) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, errors.New("brapi client requires BRAPI_API_KEY to be set")
	}

	return &Client{
		apiKey:     cfg.APIKey,
		baseURL:    strings.TrimSuffix(cfg.BaseURL, "/"),
		httpClient: &http.Client{Timeout: cfg.Timeout},
	}, nil
}

func (c *Client) FetchQuotes(ctx context.Context, tickers []string) ([]quotation.QuoteOutput, error) {
	endpoint := c.baseURL + "/api/v2/stocks/quote"
	query := url.Values{"symbols": []string{strings.Join(tickers, ",")}}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to build request: %v", quotation.ErrQuotesFetchFailed, err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: request to brapi failed: %v", quotation.ErrQuotesFetchFailed, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: brapi responded with status %d", quotation.ErrQuotesFetchFailed, resp.StatusCode)
	}

	var payload quoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, fmt.Errorf("%w: failed to decode brapi response: %v", quotation.ErrQuotesFetchFailed, err)
	}

	return mapQuotes(payload.Results), nil
}

func mapQuotes(results []quoteResult) []quotation.QuoteOutput {
	quotes := make([]quotation.QuoteOutput, 0, len(results))
	for _, result := range results {
		quotes = append(quotes, quotation.QuoteOutput{
			Ticker: result.RequestedSymbol,
			Price:  priceInCents(result.Data.RegularMarketPrice),
		})
	}
	return quotes
}

func priceInCents(price float64) int64 {
	return int64(math.Round(price * 100))
}
