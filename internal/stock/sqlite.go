package stock

import (
	"context"

	"gorm.io/gorm"
)

type sqliteRepository struct {
	db *gorm.DB
}

func NewSQLiteRepository(db *gorm.DB) Repository {
	return &sqliteRepository{db: db}
}

func (r *sqliteRepository) List(ctx context.Context) ([]Stock, error) {
	var stocks []Stock
	err := r.db.WithContext(ctx).Find(&stocks).Error
	if err != nil {
		return nil, ErrFailedToList
	}
	return stocks, nil
}

func (r *sqliteRepository) FindByTicker(ctx context.Context, ticker string) (*Stock, error) {
	var stock Stock
	err := r.db.WithContext(ctx).Where("ticker = ?", ticker).First(&stock).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, ErrStockNotFound
		}
		return nil, ErrFailedToFind
	}
	return &stock, nil
}

func (r *sqliteRepository) Create(ctx context.Context, stock *Stock) error {
	return r.db.WithContext(ctx).Create(stock).Error
}

func (r *sqliteRepository) Seed(ctx context.Context, stocks []Stock, force bool) (int, int, int, error) {
	inserted := 0
	updated := 0
	skipped := 0

	for _, stock := range stocks {
		var existing Stock
		err := r.db.WithContext(ctx).Where("ticker = ?", stock.Ticker).First(&existing).Error

		if err == gorm.ErrRecordNotFound {
			if err := r.Create(ctx, &stock); err != nil {
				return 0, 0, 0, ErrFailedToCreate
			}
			inserted++
		} else if err != nil {
			return 0, 0, 0, ErrFailedToSeed
		} else {
			if force {
				stock.ID = existing.ID
				if err := r.db.WithContext(ctx).Save(&stock).Error; err != nil {
					return 0, 0, 0, ErrFailedToSeed
				}
				updated++
			} else {
				skipped++
			}
		}
	}

	return inserted, updated, skipped, nil
}
