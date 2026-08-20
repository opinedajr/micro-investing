package stock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SQLiteRepository struct {
	db *gorm.DB
}

func NewSQLiteRepository(db *gorm.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Create(ctx context.Context, stock *Stock) error {
	return r.db.WithContext(ctx).Create(stock).Error
}

func (r *SQLiteRepository) FindByTicker(ctx context.Context, ticker string) (*Stock, error) {
	var stock Stock
	err := r.db.WithContext(ctx).Where("ticker = ?", ticker).First(&stock).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStockNotFound
		}
		return nil, err
	}
	return &stock, nil
}

func (r *SQLiteRepository) FindByID(ctx context.Context, id string) (*Stock, error) {
	var stock Stock
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&stock).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStockNotFound
		}
		return nil, err
	}
	return &stock, nil
}

func (r *SQLiteRepository) List(ctx context.Context) ([]Stock, error) {
	var stocks []Stock
	err := r.db.WithContext(ctx).Order("ticker ASC").Find(&stocks).Error
	return stocks, err
}

func (r *SQLiteRepository) Seed(ctx context.Context, stocks []Stock, force bool) error {
	if len(stocks) == 0 {
		return nil
	}

	now := time.Now()
	for i := range stocks {
		if stocks[i].ID == "" {
			stocks[i].ID = uuid.New().String()
		}
		stocks[i].CreatedAt = now
		stocks[i].UpdatedAt = now
	}

	conflictClause := "DO NOTHING"
	if force {
		conflictClause = `DO UPDATE SET
			name = EXCLUDED.name,
			sector = EXCLUDED.sector,
			rank = EXCLUDED.rank,
			website = EXCLUDED.website,
			updated_at = EXCLUDED.updated_at`
	}

	sql := fmt.Sprintf(`
		INSERT INTO stocks (id, ticker, name, sector, rank, website, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(ticker) %s`, conflictClause)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range stocks {
			stock := &stocks[i]
			website := ""
			if stock.Website != nil {
				website = *stock.Website
			}
			if err := tx.Exec(sql,
				stock.ID,
				stock.Ticker,
				stock.Name,
				stock.Sector,
				stock.Rank,
				website,
				stock.CreatedAt,
				stock.UpdatedAt,
			).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
