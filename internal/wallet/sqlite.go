package wallet

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

func (r *SQLiteRepository) Create(ctx context.Context, wallet *Wallet) error {
	return r.db.WithContext(ctx).Create(wallet).Error
}

func (r *SQLiteRepository) FindAllByUserID(ctx context.Context, userID string) ([]Wallet, error) {
	var wallets []Wallet
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&wallets).Error
	return wallets, err
}

func (r *SQLiteRepository) FindByID(ctx context.Context, id string) (*Wallet, error) {
	var wallet Wallet
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&wallet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *SQLiteRepository) FindByNameAndUserID(ctx context.Context, name string, userID string) (*Wallet, error) {
	var wallet Wallet
	err := r.db.WithContext(ctx).Where("user_id = ? AND name = ?", userID, name).First(&wallet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &wallet, nil
}

func (r *SQLiteRepository) Update(ctx context.Context, wallet *Wallet) error {
	return r.db.WithContext(ctx).Save(wallet).Error
}

func (r *SQLiteRepository) Delete(ctx context.Context, id string) error {
	var wallet Wallet
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&wallet).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("%w: %w", ErrWalletNotFound, err)
		}
		return err
	}

	return r.db.WithContext(ctx).Delete(&wallet).Error
}
