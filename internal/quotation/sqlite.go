package quotation

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type SQLiteRepository struct {
	db *gorm.DB
}

func NewSQLiteRepository(db *gorm.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) UpsertCurrentPrices(ctx context.Context, prices []CurrentPrice) error {
	if len(prices) == 0 {
		return nil
	}

	now := time.Now()
	sql := `
		INSERT INTO stocks_current_prices (stock_id, price, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(stock_id) DO UPDATE SET
			price = EXCLUDED.price,
			updated_at = EXCLUDED.updated_at`

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range prices {
			if err := tx.Exec(sql, prices[i].StockID, prices[i].Price, now).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
