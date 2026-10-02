package dividends

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

type SQLiteRepository struct {
	db *gorm.DB
}

func NewSQLiteRepository(db *gorm.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Create(ctx context.Context, dividend *Dividend) error {
	return r.db.WithContext(ctx).Create(dividend).Error
}

func (r *SQLiteRepository) FindByFilter(ctx context.Context, filter DividendFilter) ([]Dividend, error) {
	var dividends []Dividend
	query := r.db.WithContext(ctx).Where("wallet_id = ?", filter.WalletID)
	if filter.Year > 0 {
		query = query.Where("year = ?", filter.Year)
	}
	err := query.Order("year DESC").Find(&dividends).Error
	return dividends, err
}

func (r *SQLiteRepository) FindByWalletYear(ctx context.Context, walletID string, year int) (*Dividend, error) {
	var dividend Dividend
	err := r.db.WithContext(ctx).Where("wallet_id = ? AND year = ?", walletID, year).First(&dividend).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDividendNotFound
		}
		return nil, err
	}
	return &dividend, nil
}
