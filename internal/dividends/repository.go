package dividends

import "context"

type Repository interface {
	Create(ctx context.Context, dividend *Dividend) error
	FindByFilter(ctx context.Context, filter DividendFilter) ([]Dividend, error)
	FindByWalletYear(ctx context.Context, walletID string, year int) (*Dividend, error)
}
