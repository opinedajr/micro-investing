package patrimony

import "context"

type PatrimonyRepository interface {
	Create(ctx context.Context, patrimony *Patrimony) error
	Update(ctx context.Context, patrimony *Patrimony) error
	FindByID(ctx context.Context, id string) (*Patrimony, error)
	FindByFilter(ctx context.Context, filter PatrimonyFilter) ([]Patrimony, error)
	FindByWalletYearMonthType(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error)
}
