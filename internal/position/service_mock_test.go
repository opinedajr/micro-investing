package position

import (
	"context"
	"errors"

	"github.com/opinedajr/micro-investing/internal/stock"
)

type mockPositionRepository struct {
	createFunc                func(ctx context.Context, position *Position) error
	updateFunc                func(ctx context.Context, position *Position) error
	findByIDFunc              func(ctx context.Context, id string) (*Position, error)
	findByFilterFunc          func(ctx context.Context, filter PositionFilter) ([]Position, error)
	findByWalletAndStockIDFunc func(ctx context.Context, walletID string, stockID string) (*Position, error)
	findCurrentPricesMapFunc  func(ctx context.Context, stockIDs []string) (map[string]int64, error)
	sumInvestedByWalletFunc   func(ctx context.Context, walletID string) (int64, error)
	runInTransactionFunc      func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockPositionRepository) Create(ctx context.Context, position *Position) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, position)
	}
	return nil
}

func (m *mockPositionRepository) Update(ctx context.Context, position *Position) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, position)
	}
	return nil
}

func (m *mockPositionRepository) FindByID(ctx context.Context, id string) (*Position, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, ErrPositionNotFound
}

func (m *mockPositionRepository) FindByFilter(ctx context.Context, filter PositionFilter) ([]Position, error) {
	if m.findByFilterFunc != nil {
		return m.findByFilterFunc(ctx, filter)
	}
	return nil, nil
}

func (m *mockPositionRepository) FindByWalletAndStockID(ctx context.Context, walletID string, stockID string) (*Position, error) {
	if m.findByWalletAndStockIDFunc != nil {
		return m.findByWalletAndStockIDFunc(ctx, walletID, stockID)
	}
	return nil, ErrPositionNotFound
}

func (m *mockPositionRepository) FindCurrentPricesMap(ctx context.Context, stockIDs []string) (map[string]int64, error) {
	if m.findCurrentPricesMapFunc != nil {
		return m.findCurrentPricesMapFunc(ctx, stockIDs)
	}
	return map[string]int64{}, nil
}

func (m *mockPositionRepository) SumInvestedByWallet(ctx context.Context, walletID string) (int64, error) {
	if m.sumInvestedByWalletFunc != nil {
		return m.sumInvestedByWalletFunc(ctx, walletID)
	}
	return 0, nil
}

func (m *mockPositionRepository) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.runInTransactionFunc != nil {
		return m.runInTransactionFunc(ctx, fn)
	}
	return fn(ctx)
}

type mockStockRepository struct {
	findByIDFunc func(ctx context.Context, id string) (*stock.Stock, error)
}

func (m *mockStockRepository) Create(ctx context.Context, s *stock.Stock) error {
	return errors.New("not implemented")
}

func (m *mockStockRepository) FindByTicker(ctx context.Context, ticker string) (*stock.Stock, error) {
	return nil, errors.New("not implemented")
}

func (m *mockStockRepository) FindByID(ctx context.Context, id string) (*stock.Stock, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, stock.ErrStockNotFound
}

func (m *mockStockRepository) List(ctx context.Context) ([]stock.Stock, error) {
	return nil, errors.New("not implemented")
}

func (m *mockStockRepository) Seed(ctx context.Context, stocks []stock.Stock, force bool) error {
	return errors.New("not implemented")
}
