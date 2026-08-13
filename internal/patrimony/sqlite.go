package patrimony

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type SQLitePatrimonyRepository struct {
	db *gorm.DB
}

func NewSQLitePatrimonyRepository(db *gorm.DB) *SQLitePatrimonyRepository {
	return &SQLitePatrimonyRepository{db: db}
}

func (r *SQLitePatrimonyRepository) Create(ctx context.Context, patrimony *Patrimony) error {
	return r.txFromContext(ctx).Create(patrimony).Error
}

func (r *SQLitePatrimonyRepository) Update(ctx context.Context, patrimony *Patrimony) error {
	return r.txFromContext(ctx).Save(patrimony).Error
}

func (r *SQLitePatrimonyRepository) FindByID(ctx context.Context, id string) (*Patrimony, error) {
	var patrimony Patrimony
	err := r.txFromContext(ctx).Where("id = ?", id).First(&patrimony).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %w", ErrPatrimonyNotFound, err)
		}
		return nil, err
	}
	return &patrimony, nil
}

func (r *SQLitePatrimonyRepository) FindByFilter(ctx context.Context, filter PatrimonyFilter) ([]Patrimony, error) {
	var patrimonies []Patrimony
	query := r.txFromContext(ctx).Where("wallet_id = ?", filter.WalletID)

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

func (r *SQLitePatrimonyRepository) FindByWalletYearMonthType(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error) {
	var patrimony Patrimony
	err := r.txFromContext(ctx).Where("wallet_id = ? AND year = ? AND month = ? AND type = ?", walletID, year, month, assetType).First(&patrimony).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPatrimonyNotFound
		}
		return nil, err
	}
	return &patrimony, nil
}

func (r *SQLitePatrimonyRepository) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(injectTxIntoContext(ctx, tx))
	})
}

func (r *SQLitePatrimonyRepository) txFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

type SQLiteAssetRepository struct {
	db *gorm.DB
}

func NewSQLiteAssetRepository(db *gorm.DB) *SQLiteAssetRepository {
	return &SQLiteAssetRepository{db: db}
}

func (r *SQLiteAssetRepository) Create(ctx context.Context, asset *Asset) error {
	return r.txFromContext(ctx).Create(asset).Error
}

func (r *SQLiteAssetRepository) Update(ctx context.Context, asset *Asset) error {
	return r.txFromContext(ctx).Save(asset).Error
}

func (r *SQLiteAssetRepository) Delete(ctx context.Context, id string) error {
	tx := r.txFromContext(ctx)
	result := tx.Where("id = ?", id).Delete(&Asset{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrAssetNotFound
	}
	return nil
}

func (r *SQLiteAssetRepository) FindByID(ctx context.Context, id string) (*Asset, error) {
	var asset Asset
	err := r.txFromContext(ctx).Where("id = ?", id).First(&asset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}
	return &asset, nil
}

func (r *SQLiteAssetRepository) FindByFilter(ctx context.Context, filter AssetFilter) ([]Asset, error) {
	var assets []Asset
	query := r.txFromContext(ctx).Where("wallet_id = ?", filter.WalletID)
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.StartDate != nil {
		query = query.Where("date >= ?", filter.StartDate)
	}
	if filter.EndDate != nil {
		query = query.Where("date <= ?", filter.EndDate)
	}
	err := query.Order("date DESC").Find(&assets).Error
	return assets, err
}

func (r *SQLiteAssetRepository) SumByWalletTypeAndMonth(ctx context.Context, walletID string, assetType AssetType, year int, month int) (int64, error) {
	var total int64
	err := r.txFromContext(ctx).
		Model(&Asset{}).
		Where("wallet_id = ? AND type = ? AND CAST(strftime('%Y', date) AS INTEGER) = ? AND CAST(strftime('%m', date) AS INTEGER) = ?", walletID, assetType, year, month).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&total).Error
	return total, err
}

func (r *SQLiteAssetRepository) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(injectTxIntoContext(ctx, tx))
	})
}

func (r *SQLiteAssetRepository) txFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

type txKey struct{}

func injectTxIntoContext(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
