package stock

import "context"

type Repository interface {
	Create(ctx context.Context, stock *Stock) error
	FindByTicker(ctx context.Context, ticker string) (*Stock, error)
	FindByID(ctx context.Context, id string) (*Stock, error)
	FindByIDs(ctx context.Context, ids []string) ([]Stock, error)
	List(ctx context.Context) ([]Stock, error)
	Seed(ctx context.Context, stocks []Stock, force bool) error
}
