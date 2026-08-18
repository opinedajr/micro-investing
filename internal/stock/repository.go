package stock

import "context"

type Repository interface {
	List(ctx context.Context) ([]Stock, error)
	FindByTicker(ctx context.Context, ticker string) (*Stock, error)
	Create(ctx context.Context, stock *Stock) error
	Seed(ctx context.Context, stocks []Stock, force bool) (inserted, updated, skipped int, err error)
}
