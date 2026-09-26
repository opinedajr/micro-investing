package quotation

import "context"

type Provider interface {
	FetchQuotes(ctx context.Context, tickers []string) ([]QuoteOutput, error)
}
