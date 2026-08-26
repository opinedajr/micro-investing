package position

import "context"

type Repository interface {
	Create(ctx context.Context, position *Position) error
	Update(ctx context.Context, position *Position) error
	FindByID(ctx context.Context, id string) (*Position, error)
	FindByFilter(ctx context.Context, filter PositionFilter) ([]Position, error)
	FindByWalletAndStockID(ctx context.Context, walletID string, stockID string) (*Position, error)
	FindCurrentPricesMap(ctx context.Context, stockIDs []string) (map[string]int64, error)
	SumInvestedByWallet(ctx context.Context, walletID string) (int64, error)
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
