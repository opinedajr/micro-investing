package dashboard

import (
	"context"
	"errors"
	"testing"

	"github.com/opinedajr/micro-investing/internal/patrimony"
	"github.com/opinedajr/micro-investing/internal/position"
	"github.com/opinedajr/micro-investing/internal/stock"
	"github.com/stretchr/testify/assert"
)

type mockPatrimonyRepository struct {
	findLatestMonthByWalletFunc func(ctx context.Context, walletID string) (int, int, error)
	sumByWalletYearMonthFunc    func(ctx context.Context, walletID string, year int, month int) ([]patrimony.TypeAmount, error)
}

func (m *mockPatrimonyRepository) Create(ctx context.Context, p *patrimony.Patrimony) error {
	return nil
}

func (m *mockPatrimonyRepository) Update(ctx context.Context, p *patrimony.Patrimony) error {
	return nil
}

func (m *mockPatrimonyRepository) FindByID(ctx context.Context, id string) (*patrimony.Patrimony, error) {
	return nil, nil
}

func (m *mockPatrimonyRepository) FindByFilter(ctx context.Context, filter patrimony.PatrimonyFilter) ([]patrimony.Patrimony, error) {
	return nil, nil
}

func (m *mockPatrimonyRepository) FindByWalletYearMonthType(ctx context.Context, walletID string, year int, month int, assetType patrimony.AssetType) (*patrimony.Patrimony, error) {
	return nil, nil
}

func (m *mockPatrimonyRepository) FindLatestMonthByWallet(ctx context.Context, walletID string) (int, int, error) {
	if m.findLatestMonthByWalletFunc != nil {
		return m.findLatestMonthByWalletFunc(ctx, walletID)
	}
	return 0, 0, nil
}

func (m *mockPatrimonyRepository) SumByWalletYearMonth(ctx context.Context, walletID string, year int, month int) ([]patrimony.TypeAmount, error) {
	if m.sumByWalletYearMonthFunc != nil {
		return m.sumByWalletYearMonthFunc(ctx, walletID, year, month)
	}
	return nil, nil
}

func (m *mockPatrimonyRepository) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type mockPositionRepository struct {
	sumInvestedByWalletFunc func(ctx context.Context, walletID string) (int64, error)
}

func (m *mockPositionRepository) Create(ctx context.Context, p *position.Position) error {
	return nil
}

func (m *mockPositionRepository) Update(ctx context.Context, p *position.Position) error {
	return nil
}

func (m *mockPositionRepository) FindByID(ctx context.Context, id string) (*position.Position, error) {
	return nil, nil
}

func (m *mockPositionRepository) FindByFilter(ctx context.Context, filter position.PositionFilter) ([]position.Position, error) {
	return nil, nil
}

func (m *mockPositionRepository) FindByWalletAndStockID(ctx context.Context, walletID string, stockID string) (*position.Position, error) {
	return nil, nil
}

func (m *mockPositionRepository) FindCurrentPricesMap(ctx context.Context, stockIDs []string) (map[string]int64, error) {
	return nil, nil
}

func (m *mockPositionRepository) SumInvestedByWallet(ctx context.Context, walletID string) (int64, error) {
	if m.sumInvestedByWalletFunc != nil {
		return m.sumInvestedByWalletFunc(ctx, walletID)
	}
	return 0, nil
}

func (m *mockPositionRepository) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

type mockStockRepository struct{}

func (m *mockStockRepository) Create(ctx context.Context, s *stock.Stock) error {
	return nil
}

func (m *mockStockRepository) FindByTicker(ctx context.Context, ticker string) (*stock.Stock, error) {
	return nil, nil
}

func (m *mockStockRepository) FindByID(ctx context.Context, id string) (*stock.Stock, error) {
	return nil, nil
}

func (m *mockStockRepository) List(ctx context.Context) ([]stock.Stock, error) {
	return nil, nil
}

func (m *mockStockRepository) Seed(ctx context.Context, stocks []stock.Stock, force bool) error {
	return nil
}

func TestService_Summary(t *testing.T) {
	t.Run("success - returns aggregated summary", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			findLatestMonthByWalletFunc: func(ctx context.Context, walletID string) (int, int, error) {
				return 2026, 7, nil
			},
			sumByWalletYearMonthFunc: func(ctx context.Context, walletID string, year int, month int) ([]patrimony.TypeAmount, error) {
				return []patrimony.TypeAmount{
					{Type: patrimony.TypeStocks, Amount: 500000},
					{Type: patrimony.TypeFixedIncome, Amount: 500000},
					{Type: patrimony.TypeEmergencyReserve, Amount: 250000},
					{Type: patrimony.TypeLiquidCash, Amount: 250000},
				}, nil
			},
		}
		positionRepo := &mockPositionRepository{
			sumInvestedByWalletFunc: func(ctx context.Context, walletID string) (int64, error) {
				return 500000, nil
			},
		}

		service := NewService(patrimonyRepo, positionRepo, &mockStockRepository{})
		output, err := service.Summary(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(1500000), output.CurrentPatrimony)
		assert.Equal(t, int64(500000), output.StocksInvested)
		assert.Equal(t, int64(0), output.YearlyDividends)
	})

	t.Run("success - returns zero when wallet has no patrimony records", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			findLatestMonthByWalletFunc: func(ctx context.Context, walletID string) (int, int, error) {
				return 0, 0, nil
			},
			sumByWalletYearMonthFunc: func(ctx context.Context, walletID string, year int, month int) ([]patrimony.TypeAmount, error) {
				return nil, nil
			},
		}
		positionRepo := &mockPositionRepository{
			sumInvestedByWalletFunc: func(ctx context.Context, walletID string) (int64, error) {
				return 0, nil
			},
		}

		service := NewService(patrimonyRepo, positionRepo, &mockStockRepository{})
		output, err := service.Summary(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(0), output.CurrentPatrimony)
		assert.Equal(t, int64(0), output.StocksInvested)
		assert.Equal(t, int64(0), output.YearlyDividends)
	})

	t.Run("error - returns error when finding latest month fails", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			findLatestMonthByWalletFunc: func(ctx context.Context, walletID string) (int, int, error) {
				return 0, 0, errors.New("database error")
			},
		}
		positionRepo := &mockPositionRepository{}

		service := NewService(patrimonyRepo, positionRepo, &mockStockRepository{})
		_, err := service.Summary(context.Background(), "wallet-id")

		assert.Error(t, err)
	})

	t.Run("error - returns error when summing patrimony fails", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			findLatestMonthByWalletFunc: func(ctx context.Context, walletID string) (int, int, error) {
				return 2026, 7, nil
			},
			sumByWalletYearMonthFunc: func(ctx context.Context, walletID string, year int, month int) ([]patrimony.TypeAmount, error) {
				return nil, errors.New("database error")
			},
		}
		positionRepo := &mockPositionRepository{}

		service := NewService(patrimonyRepo, positionRepo, &mockStockRepository{})
		_, err := service.Summary(context.Background(), "wallet-id")

		assert.Error(t, err)
	})

	t.Run("error - returns error when summing positions fails", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			findLatestMonthByWalletFunc: func(ctx context.Context, walletID string) (int, int, error) {
				return 2026, 7, nil
			},
			sumByWalletYearMonthFunc: func(ctx context.Context, walletID string, year int, month int) ([]patrimony.TypeAmount, error) {
				return []patrimony.TypeAmount{{Type: patrimony.TypeStocks, Amount: 500000}}, nil
			},
		}
		positionRepo := &mockPositionRepository{
			sumInvestedByWalletFunc: func(ctx context.Context, walletID string) (int64, error) {
				return 0, errors.New("database error")
			},
		}

		service := NewService(patrimonyRepo, positionRepo, &mockStockRepository{})
		_, err := service.Summary(context.Background(), "wallet-id")

		assert.Error(t, err)
	})
}
