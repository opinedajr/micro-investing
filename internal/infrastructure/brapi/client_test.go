package brapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/opinedajr/micro-investing/internal/infrastructure/brapi"
	"github.com/opinedajr/micro-investing/internal/quotation"
	"github.com/opinedajr/micro-investing/internal/shared/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type serverCapture struct {
	authHeader string
	symbols    string
}

func newTestClient(t *testing.T, handler http.HandlerFunc) (*brapi.Client, *serverCapture) {
	t.Helper()

	capture := &serverCapture{}
	wrapped := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capture.authHeader = r.Header.Get("Authorization")
		capture.symbols = r.URL.Query().Get("symbols")
		handler(w, r)
	})

	server := httptest.NewServer(wrapped)
	t.Cleanup(server.Close)

	client, err := brapi.NewClient(config.BrapiConfig{
		APIKey:  "test-api-key",
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	})
	require.NoError(t, err)

	return client, capture
}

func respondWithJSON(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	_, err := w.Write([]byte(body))
	require.NoError(t, err)
}

func TestNewClient(t *testing.T) {
	t.Run("error - missing API key fails fast", func(t *testing.T) {
		client, err := brapi.NewClient(config.BrapiConfig{
			APIKey:  "",
			BaseURL: "https://brapi.dev",
			Timeout: 15 * time.Second,
		})

		assert.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "BRAPI_API_KEY")
	})

	t.Run("success - builds client with API key", func(t *testing.T) {
		client, err := brapi.NewClient(config.BrapiConfig{
			APIKey:  "some-key",
			BaseURL: "https://brapi.dev",
			Timeout: 15 * time.Second,
		})

		assert.NoError(t, err)
		assert.NotNil(t, client)
	})
}

func TestFetchQuotes_Success(t *testing.T) {
	t.Run("success - parses payload and rounds prices to cents", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			respondWithJSON(t, w, `{
				"results": [
					{"requestedSymbol": "PETR4", "data": {"regularMarketPrice": 41.18}},
					{"requestedSymbol": "VALE3", "data": {"regularMarketPrice": 62.555}},
					{"requestedSymbol": "ITSA4", "data": {"regularMarketPrice": 8.33}}
				]
			}`)
		})

		quotes, err := client.FetchQuotes(context.Background(), []string{"PETR4", "VALE3", "ITSA4"})

		require.NoError(t, err)
		assert.Equal(t, []quotation.QuoteOutput{
			{Ticker: "PETR4", Price: 4118},
			{Ticker: "VALE3", Price: 6256},
			{Ticker: "ITSA4", Price: 833},
		}, quotes)
	})

	t.Run("success - maps requestedSymbol to ticker", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			respondWithJSON(t, w, `{
				"results": [
					{"requestedSymbol": "PETR4", "symbol": "RENAMED4", "data": {"regularMarketPrice": 10.0}}
				]
			}`)
		})

		quotes, err := client.FetchQuotes(context.Background(), []string{"PETR4"})

		require.NoError(t, err)
		require.Len(t, quotes, 1)
		assert.Equal(t, "PETR4", quotes[0].Ticker)
		assert.Equal(t, int64(1000), quotes[0].Price)
	})

	t.Run("success - missing ticker is simply not returned", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			respondWithJSON(t, w, `{
				"results": [
					{"requestedSymbol": "PETR4", "data": {"regularMarketPrice": 41.18}}
				]
			}`)
		})

		quotes, err := client.FetchQuotes(context.Background(), []string{"PETR4", "VALE3"})

		require.NoError(t, err)
		assert.Equal(t, []quotation.QuoteOutput{{Ticker: "PETR4", Price: 4118}}, quotes)
	})

	t.Run("success - empty results returns no quotes", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			respondWithJSON(t, w, `{"results": []}`)
		})

		quotes, err := client.FetchQuotes(context.Background(), []string{"PETR4"})

		require.NoError(t, err)
		assert.Empty(t, quotes)
	})
}

func TestFetchQuotes_Request(t *testing.T) {
	t.Run("success - sends bearer auth header and comma-joined symbols", func(t *testing.T) {
		client, capture := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/api/v2/stocks/quote", r.URL.Path)
			assert.Equal(t, http.MethodGet, r.Method)
			respondWithJSON(t, w, `{"results": []}`)
		})

		_, err := client.FetchQuotes(context.Background(), []string{"PETR4", "VALE3"})

		require.NoError(t, err)
		assert.Equal(t, "Bearer test-api-key", capture.authHeader)
		assert.Equal(t, "PETR4,VALE3", capture.symbols)
	})
}

func TestFetchQuotes_Errors(t *testing.T) {
	t.Run("error - unauthorized 401", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		})

		quotes, err := client.FetchQuotes(context.Background(), []string{"PETR4"})

		assert.ErrorIs(t, err, quotation.ErrQuotesFetchFailed)
		assert.Contains(t, err.Error(), "401")
		assert.Nil(t, quotes)
	})

	t.Run("error - rate limited 429", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		})

		quotes, err := client.FetchQuotes(context.Background(), []string{"PETR4"})

		assert.ErrorIs(t, err, quotation.ErrQuotesFetchFailed)
		assert.Contains(t, err.Error(), "429")
		assert.Nil(t, quotes)
	})

	t.Run("error - server error 500", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		})

		quotes, err := client.FetchQuotes(context.Background(), []string{"PETR4"})

		assert.ErrorIs(t, err, quotation.ErrQuotesFetchFailed)
		assert.Nil(t, quotes)
	})

	t.Run("error - network failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		serverURL := server.URL
		server.Close()

		client, err := brapi.NewClient(config.BrapiConfig{
			APIKey:  "test-api-key",
			BaseURL: serverURL,
			Timeout: 5 * time.Second,
		})
		require.NoError(t, err)

		quotes, err := client.FetchQuotes(context.Background(), []string{"PETR4"})

		assert.ErrorIs(t, err, quotation.ErrQuotesFetchFailed)
		assert.Nil(t, quotes)
	})

	t.Run("error - malformed json body", func(t *testing.T) {
		client, _ := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
			respondWithJSON(t, w, `{"results": [`)
		})

		quotes, err := client.FetchQuotes(context.Background(), []string{"PETR4"})

		assert.ErrorIs(t, err, quotation.ErrQuotesFetchFailed)
		assert.Nil(t, quotes)
	})
}
