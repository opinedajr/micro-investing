package patrimony

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type SQLiteRepository struct {
	db *gorm.DB
}

func NewSQLiteRepository(db *gorm.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Create(ctx context.Context, patrimony *Patrimony) error {
	return r.db.WithContext(ctx).Create(patrimony).Error
}

func (r *SQLiteRepository) Update(ctx context.Context, patrimony *Patrimony) error {
	return r.db.WithContext(ctx).Save(patrimony).Error
}

func (r *SQLiteRepository) FindByID(ctx context.Context, id string) (*Patrimony, error) {
	var patrimony Patrimony
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&patrimony).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %w", ErrPatrimonyNotFound, err)
		}
		return nil, err
	}
	return &patrimony, nil
}

func (r *SQLiteRepository) FindByFilter(ctx context.Context, filter PatrimonyFilter) ([]Patrimony, error) {
	var patrimonies []Patrimony
	query := r.db.WithContext(ctx).Where("wallet_id = ?", filter.WalletID)

	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Year > 0 {
		query = query.Where("year = ?", filter.Year)
	}
	if filter.Month > 0 {
		query = query.Where("month = ?", filter.Month)
	}

	err := query.Order("year DESC, month DESC, type").Find(&patrimonies).Error
	return patrimonies, err
}

func (r *SQLiteRepository) FindByWalletYearMonthType(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error) {
	var patrimony Patrimony
	err := r.db.WithContext(ctx).Where("wallet_id = ? AND year = ? AND month = ? AND type = ?", walletID, year, month, assetType).First(&patrimony).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPatrimonyNotFound
		}
		return nil, err
	}
	return &patrimony, nil
}
