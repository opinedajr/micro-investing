package patrimony

import "context"

type PatrimonyRepository interface {
	Create(ctx context.Context, patrimony *Patrimony) error
	Update(ctx context.Context, patrimony *Patrimony) error
	FindByID(ctx context.Context, id string) (*Patrimony, error)
	FindByFilter(ctx context.Context, filter PatrimonyFilter) ([]Patrimony, error)
	FindByWalletYearMonthType(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error)
	FindLatestMonthByWallet(ctx context.Context, walletID string) (year int, month int, err error)
	SumByWalletYearMonth(ctx context.Context, walletID string, year int, month int) ([]TypeAmount, error)
	RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type TypeAmount struct {
	Type   AssetType
	Amount int64
}

type AssetRepository interface {
	Create(ctx context.Context, asset *Asset) error
	Update(ctx context.Context, asset *Asset) error
	Delete(ctx context.Context, id string) error
	FindByID(ctx context.Context, id string) (*Asset, error)
	FindByFilter(ctx context.Context, filter AssetFilter) ([]Asset, error)
	SumByWalletTypeAndMonth(ctx context.Context, walletID string, assetType AssetType, year int, month int) (int64, error)
}
