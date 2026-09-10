package position

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

var sortColumnWhitelist = map[string]string{
	"ticker":            "stocks.ticker",
	"rank":              "stocks.rank",
	"invested":          "positions.invested",
	"variation_percent": "positions.variation_percent",
	"portfolio_percent": "positions.portfolio_percent",
}

type SQLiteRepository struct {
	db *gorm.DB
}

func NewSQLiteRepository(db *gorm.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Create(ctx context.Context, position *Position) error {
	return r.txFromContext(ctx).Create(position).Error
}

func (r *SQLiteRepository) Update(ctx context.Context, position *Position) error {
	return r.txFromContext(ctx).Save(position).Error
}

func (r *SQLiteRepository) FindByID(ctx context.Context, id string) (*Position, error) {
	var position Position
	err := r.txFromContext(ctx).Where("id = ?", id).First(&position).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: %w", ErrPositionNotFound, err)
		}
		return nil, err
	}
	return &position, nil
}

func (r *SQLiteRepository) FindByFilter(ctx context.Context, filter PositionFilter) ([]Position, error) {
	var positions []Position
	query := r.txFromContext(ctx).
		Table("positions").
		Select("positions.*").
		Joins("JOIN stocks ON stocks.id = positions.stock_id").
		Where("positions.wallet_id = ?", filter.WalletID)

	if filter.Ticker != "" {
		query = query.Where("stocks.ticker LIKE UPPER(?)", "%"+filter.Ticker+"%")
	}

	query = query.Order(r.buildOrderClause(filter.Sort))
	err := query.Find(&positions).Error
	return positions, err
}

func (r *SQLiteRepository) buildOrderClause(sort string) string {
	direction := "ASC"
	if strings.HasPrefix(sort, "-") {
		direction = "DESC"
		sort = sort[1:]
	}

	column, ok := sortColumnWhitelist[sort]
	if !ok {
		return "positions.balance DESC"
	}

	return column + " " + direction
}

func (r *SQLiteRepository) FindByWalletAndStockID(ctx context.Context, walletID string, stockID string) (*Position, error) {
	var position Position
	err := r.txFromContext(ctx).
		Where("wallet_id = ? AND stock_id = ?", walletID, stockID).
		First(&position).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPositionNotFound
		}
		return nil, err
	}
	return &position, nil
}

func (r *SQLiteRepository) FindCurrentPricesMap(ctx context.Context, stockIDs []string) (map[string]int64, error) {
	prices := make(map[string]int64)
	if len(stockIDs) == 0 {
		return prices, nil
	}

	type result struct {
		StockID string
		Price   int64
	}

	var rows []result
	err := r.txFromContext(ctx).
		Table("stocks_current_prices").
		Select("stock_id, price").
		Where("stock_id IN ?", stockIDs).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		prices[row.StockID] = row.Price
	}
	return prices, nil
}

func (r *SQLiteRepository) SumInvestedByWallet(ctx context.Context, walletID string) (int64, error) {
	var total int64
	err := r.txFromContext(ctx).
		Model(&Position{}).
		Where("wallet_id = ?", walletID).
		Select("COALESCE(SUM(invested), 0)").
		Scan(&total).Error
	return total, err
}

func (r *SQLiteRepository) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(injectTxIntoContext(ctx, tx))
	})
}

func (r *SQLiteRepository) txFromContext(ctx context.Context) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return r.db.WithContext(ctx)
}

type txKey struct{}

func injectTxIntoContext(ctx context.Context, tx *gorm.DB) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}
