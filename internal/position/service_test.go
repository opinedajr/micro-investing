package position

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/opinedajr/micro-investing/internal/shared/logger"
	"github.com/opinedajr/micro-investing/internal/stock"
	"github.com/stretchr/testify/assert"
)

type noopLogger struct{}

func (n noopLogger) Debug(ctx context.Context, msg string, args ...any)   {}
func (n noopLogger) Info(ctx context.Context, msg string, args ...any)    {}
func (n noopLogger) Warn(ctx context.Context, msg string, args ...any)    {}
func (n noopLogger) Error(ctx context.Context, msg string, args ...any)   {}
func (n noopLogger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
}
func (n noopLogger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
}
func (n noopLogger) With(args ...any) logger.Logger      { return n }
func (n noopLogger) WithGroup(name string) logger.Logger { return n }

type memoryPositionRepository struct {
	positions map[string]*Position
	prices    map[string]int64
}

func newMemoryPositionRepository() *memoryPositionRepository {
	return &memoryPositionRepository{
		positions: make(map[string]*Position),
		prices:    make(map[string]int64),
	}
}

func (m *memoryPositionRepository) Create(ctx context.Context, position *Position) error {
	m.positions[position.ID] = position
	return nil
}

func (m *memoryPositionRepository) Update(ctx context.Context, position *Position) error {
	m.positions[position.ID] = position
	return nil
}

func (m *memoryPositionRepository) FindByID(ctx context.Context, id string) (*Position, error) {
	p, ok := m.positions[id]
	if !ok {
		return nil, ErrPositionNotFound
	}
	return p, nil
}

func (m *memoryPositionRepository) FindByFilter(ctx context.Context, filter PositionFilter) ([]Position, error) {
	var result []Position
	for _, p := range m.positions {
		if p.WalletID == filter.WalletID {
			result = append(result, *p)
		}
	}
	return result, nil
}

func (m *memoryPositionRepository) FindByWalletAndStockID(ctx context.Context, walletID string, stockID string) (*Position, error) {
	for _, p := range m.positions {
		if p.WalletID == walletID && p.StockID == stockID {
			return p, nil
		}
	}
	return nil, ErrPositionNotFound
}

func (m *memoryPositionRepository) FindCurrentPricesMap(ctx context.Context, stockIDs []string) (map[string]int64, error) {
	result := make(map[string]int64)
	for _, id := range stockIDs {
		if price, ok := m.prices[id]; ok {
			result[id] = price
		}
	}
	return result, nil
}

func (m *memoryPositionRepository) SumInvestedByWallet(ctx context.Context, walletID string) (int64, error) {
	var total int64
	for _, p := range m.positions {
		if p.WalletID == walletID {
			total += p.Invested
		}
	}
	return total, nil
}

func (m *memoryPositionRepository) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestService_Create(t *testing.T) {
	t.Run("success - creates position and consolidates wallet", func(t *testing.T) {
		stockRepo := &mockStockRepository{
			findByIDFunc: func(ctx context.Context, id string) (*stock.Stock, error) {
				return &stock.Stock{ID: id, Ticker: "PETR4"}, nil
			},
		}
		positionRepo := newMemoryPositionRepository()
		positionRepo.prices["stock-id"] = 7500

		service := NewService(positionRepo, stockRepo, noopLogger{})
		input := CreatePositionInput{
			WalletID:     "wallet-id",
			StockID:      "stock-id",
			Quantity:     100,
			AveragePrice: 5000,
		}

		output, err := service.Create(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "wallet-id", output.WalletID)
		assert.Equal(t, "stock-id", output.StockID)
		assert.Equal(t, int64(100), output.Quantity)
		assert.Equal(t, int64(5000), output.AveragePrice)
		assert.Equal(t, int64(7500), output.CurrentPrice)
		assert.Equal(t, int64(500000), output.Invested)
		assert.Equal(t, int64(750000), output.Balance)
		assert.Equal(t, int64(250000), output.VariationValue)
		assert.InDelta(t, 50.0, output.VariationPercent, 0.001)
		assert.InDelta(t, 100.0, output.PortfolioPercent, 0.001)
	})

	t.Run("success - consolidates existing positions when adding new one", func(t *testing.T) {
		stockRepo := &mockStockRepository{
			findByIDFunc: func(ctx context.Context, id string) (*stock.Stock, error) {
				return &stock.Stock{ID: id, Ticker: "VALE3"}, nil
			},
		}
		positionRepo := newMemoryPositionRepository()
		positionRepo.prices["s1"] = 1000
		positionRepo.prices["s2"] = 1000
		positionRepo.positions["p1"] = &Position{
			ID: "p1", WalletID: "wallet-id", StockID: "s1", Quantity: 100, AveragePrice: 1000,
		}

		service := NewService(positionRepo, stockRepo, noopLogger{})
		input := CreatePositionInput{
			WalletID:     "wallet-id",
			StockID:      "s2",
			Quantity:     100,
			AveragePrice: 1000,
		}

		output, err := service.Create(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.InDelta(t, 50.0, positionRepo.positions["p1"].PortfolioPercent, 0.001)
		assert.InDelta(t, 50.0, positionRepo.positions[output.ID].PortfolioPercent, 0.001)
	})

	t.Run("error - returns stock not found when stock does not exist", func(t *testing.T) {
		stockRepo := &mockStockRepository{
			findByIDFunc: func(ctx context.Context, id string) (*stock.Stock, error) {
				return nil, stock.ErrStockNotFound
			},
		}
		positionRepo := newMemoryPositionRepository()

		service := NewService(positionRepo, stockRepo, noopLogger{})
		input := CreatePositionInput{
			WalletID:     "wallet-id",
			StockID:      "missing-stock",
			Quantity:     10,
			AveragePrice: 1000,
		}

		_, err := service.Create(context.Background(), input)

		assert.Error(t, err)
		assert.Equal(t, ErrStockNotFound, err)
	})

	t.Run("error - returns position already exists for duplicate wallet/stock", func(t *testing.T) {
		stockRepo := &mockStockRepository{
			findByIDFunc: func(ctx context.Context, id string) (*stock.Stock, error) {
				return &stock.Stock{ID: id, Ticker: "PETR4"}, nil
			},
		}
		positionRepo := newMemoryPositionRepository()
		positionRepo.positions["existing"] = &Position{
			ID: "existing", WalletID: "wallet-id", StockID: "stock-id", Quantity: 10, AveragePrice: 1000,
		}

		service := NewService(positionRepo, stockRepo, noopLogger{})
		input := CreatePositionInput{
			WalletID:     "wallet-id",
			StockID:      "stock-id",
			Quantity:     10,
			AveragePrice: 1000,
		}

		_, err := service.Create(context.Background(), input)

		assert.Error(t, err)
		assert.Equal(t, ErrPositionAlreadyExists, err)
	})

	t.Run("error - returns validation error for invalid quantity", func(t *testing.T) {
		service := NewService(newMemoryPositionRepository(), &mockStockRepository{}, noopLogger{})
		input := CreatePositionInput{
			WalletID:     "wallet-id",
			StockID:      "stock-id",
			Quantity:     0,
			AveragePrice: 1000,
		}

		_, err := service.Create(context.Background(), input)

		assert.Error(t, err)
	})

	t.Run("error - returns validation error for invalid average_price", func(t *testing.T) {
		service := NewService(newMemoryPositionRepository(), &mockStockRepository{}, noopLogger{})
		input := CreatePositionInput{
			WalletID:     "wallet-id",
			StockID:      "stock-id",
			Quantity:     10,
			AveragePrice: 0,
		}

		_, err := service.Create(context.Background(), input)

		assert.Error(t, err)
	})

	t.Run("error - returns error when consolidation fails", func(t *testing.T) {
		stockRepo := &mockStockRepository{
			findByIDFunc: func(ctx context.Context, id string) (*stock.Stock, error) {
				return &stock.Stock{ID: id, Ticker: "PETR4"}, nil
			},
		}
		positionRepo := &mockPositionRepository{
			findByWalletAndStockIDFunc: func(ctx context.Context, walletID string, stockID string) (*Position, error) {
				return nil, ErrPositionNotFound
			},
			runInTransactionFunc: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
			createFunc: func(ctx context.Context, position *Position) error {
				return nil
			},
			findByFilterFunc: func(ctx context.Context, filter PositionFilter) ([]Position, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(positionRepo, stockRepo, noopLogger{})
		input := CreatePositionInput{
			WalletID:     "wallet-id",
			StockID:      "stock-id",
			Quantity:     10,
			AveragePrice: 1000,
		}

		_, err := service.Create(context.Background(), input)

		assert.Error(t, err)
	})
}

func TestService_List(t *testing.T) {
	t.Run("success - returns positions for wallet ordered by balance desc", func(t *testing.T) {
		positionRepo := newMemoryPositionRepository()
		positionRepo.positions["p1"] = &Position{ID: "p1", WalletID: "wallet-id", StockID: "s1", Balance: 10000, Invested: 10000, PortfolioPercent: 100}
		positionRepo.positions["p2"] = &Position{ID: "p2", WalletID: "wallet-id", StockID: "s2", Balance: 30000, Invested: 20000, PortfolioPercent: 100}
		positionRepo.positions["p3"] = &Position{ID: "p3", WalletID: "other-wallet", StockID: "s3", Balance: 50000, Invested: 50000, PortfolioPercent: 100}

		service := NewService(positionRepo, &mockStockRepository{}, noopLogger{})
		outputs, err := service.List(context.Background(), PositionFilter{WalletID: "wallet-id"})

		assert.NoError(t, err)
		assert.Len(t, outputs, 2)
		assert.ElementsMatch(t, []string{"p1", "p2"}, []string{outputs[0].ID, outputs[1].ID})
	})

	t.Run("success - returns empty slice when no positions", func(t *testing.T) {
		service := NewService(newMemoryPositionRepository(), &mockStockRepository{}, noopLogger{})
		outputs, err := service.List(context.Background(), PositionFilter{WalletID: "wallet-id"})

		assert.NoError(t, err)
		assert.NotNil(t, outputs)
		assert.Empty(t, outputs)
	})

	t.Run("error - returns error when repository fails", func(t *testing.T) {
		positionRepo := &mockPositionRepository{
			findByFilterFunc: func(ctx context.Context, filter PositionFilter) ([]Position, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(positionRepo, &mockStockRepository{}, noopLogger{})
		_, err := service.List(context.Background(), PositionFilter{WalletID: "wallet-id"})

		assert.Error(t, err)
	})
}

func TestService_Find(t *testing.T) {
	t.Run("success - returns position when it belongs to wallet", func(t *testing.T) {
		positionRepo := newMemoryPositionRepository()
		positionRepo.positions["p1"] = &Position{ID: "p1", WalletID: "wallet-id", StockID: "s1", Balance: 10000, Invested: 10000}

		service := NewService(positionRepo, &mockStockRepository{}, noopLogger{})
		output, err := service.Find(context.Background(), "wallet-id", "p1")

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "p1", output.ID)
		assert.Equal(t, "wallet-id", output.WalletID)
	})

	t.Run("error - returns not found when position belongs to another wallet", func(t *testing.T) {
		positionRepo := newMemoryPositionRepository()
		positionRepo.positions["p1"] = &Position{ID: "p1", WalletID: "other-wallet", StockID: "s1", Balance: 10000, Invested: 10000}

		service := NewService(positionRepo, &mockStockRepository{}, noopLogger{})
		_, err := service.Find(context.Background(), "wallet-id", "p1")

		assert.ErrorIs(t, err, ErrPositionNotFound)
	})

	t.Run("error - returns not found when position does not exist", func(t *testing.T) {
		service := NewService(newMemoryPositionRepository(), &mockStockRepository{}, noopLogger{})
		_, err := service.Find(context.Background(), "wallet-id", "missing-id")

		assert.ErrorIs(t, err, ErrPositionNotFound)
	})
}

func TestService_Update(t *testing.T) {
	t.Run("success - overwrites quantity and average_price", func(t *testing.T) {
		positionRepo := newMemoryPositionRepository()
		positionRepo.positions["p1"] = &Position{
			ID: "p1", WalletID: "wallet-id", StockID: "s1", Quantity: 10, AveragePrice: 1000,
		}
		positionRepo.prices["s1"] = 2000

		service := NewService(positionRepo, &mockStockRepository{}, noopLogger{})
		output, err := service.Update(context.Background(), UpdatePositionInput{
			WalletID:     "wallet-id",
			PositionID:   "p1",
			Quantity:     50,
			AveragePrice: 1500,
		})

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, int64(50), output.Quantity)
		assert.Equal(t, int64(1500), output.AveragePrice)
		assert.Equal(t, int64(75000), output.Invested)
		assert.Equal(t, int64(100000), output.Balance)
	})

	t.Run("success - redistributes portfolio_percent across wallet positions", func(t *testing.T) {
		positionRepo := newMemoryPositionRepository()
		positionRepo.positions["p1"] = &Position{
			ID: "p1", WalletID: "wallet-id", StockID: "s1", Quantity: 100, AveragePrice: 1000,
		}
		positionRepo.positions["p2"] = &Position{
			ID: "p2", WalletID: "wallet-id", StockID: "s2", Quantity: 100, AveragePrice: 1000,
		}
		positionRepo.prices["s1"] = 1000
		positionRepo.prices["s2"] = 1000

		service := NewService(positionRepo, &mockStockRepository{}, noopLogger{})
		_, err := service.Update(context.Background(), UpdatePositionInput{
			WalletID:     "wallet-id",
			PositionID:   "p1",
			Quantity:     200,
			AveragePrice: 1000,
		})

		assert.NoError(t, err)
		assert.InDelta(t, 66.6667, positionRepo.positions["p1"].PortfolioPercent, 0.001)
		assert.InDelta(t, 33.3333, positionRepo.positions["p2"].PortfolioPercent, 0.001)
	})

	t.Run("success - keeps stock_id immutable", func(t *testing.T) {
		positionRepo := newMemoryPositionRepository()
		positionRepo.positions["p1"] = &Position{
			ID: "p1", WalletID: "wallet-id", StockID: "s1", Quantity: 10, AveragePrice: 1000,
		}
		positionRepo.prices["s1"] = 1000

		service := NewService(positionRepo, &mockStockRepository{}, noopLogger{})
		output, err := service.Update(context.Background(), UpdatePositionInput{
			WalletID:     "wallet-id",
			PositionID:   "p1",
			Quantity:     20,
			AveragePrice: 2000,
		})

		assert.NoError(t, err)
		assert.Equal(t, "s1", output.StockID)
	})

	t.Run("error - returns not found when position belongs to another wallet", func(t *testing.T) {
		positionRepo := newMemoryPositionRepository()
		positionRepo.positions["p1"] = &Position{
			ID: "p1", WalletID: "other-wallet", StockID: "s1", Quantity: 10, AveragePrice: 1000,
		}

		service := NewService(positionRepo, &mockStockRepository{}, noopLogger{})
		_, err := service.Update(context.Background(), UpdatePositionInput{
			WalletID:     "wallet-id",
			PositionID:   "p1",
			Quantity:     20,
			AveragePrice: 2000,
		})

		assert.ErrorIs(t, err, ErrPositionNotFound)
	})

	t.Run("error - returns not found when position does not exist", func(t *testing.T) {
		service := NewService(newMemoryPositionRepository(), &mockStockRepository{}, noopLogger{})
		_, err := service.Update(context.Background(), UpdatePositionInput{
			WalletID:     "wallet-id",
			PositionID:   "missing-id",
			Quantity:     20,
			AveragePrice: 2000,
		})

		assert.ErrorIs(t, err, ErrPositionNotFound)
	})

	t.Run("error - returns validation error for invalid quantity", func(t *testing.T) {
		service := NewService(newMemoryPositionRepository(), &mockStockRepository{}, noopLogger{})
		_, err := service.Update(context.Background(), UpdatePositionInput{
			WalletID:     "wallet-id",
			PositionID:   "p1",
			Quantity:     0,
			AveragePrice: 1000,
		})

		assert.Error(t, err)
	})

	t.Run("error - returns validation error for invalid average_price", func(t *testing.T) {
		service := NewService(newMemoryPositionRepository(), &mockStockRepository{}, noopLogger{})
		_, err := service.Update(context.Background(), UpdatePositionInput{
			WalletID:     "wallet-id",
			PositionID:   "p1",
			Quantity:     10,
			AveragePrice: 0,
		})

		assert.Error(t, err)
	})

	t.Run("error - returns error when consolidation fails", func(t *testing.T) {
		positionRepo := &mockPositionRepository{
			findByIDFunc: func(ctx context.Context, id string) (*Position, error) {
				return &Position{ID: id, WalletID: "wallet-id", StockID: "s1", Quantity: 10, AveragePrice: 1000}, nil
			},
			runInTransactionFunc: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
			updateFunc: func(ctx context.Context, position *Position) error {
				return nil
			},
			findByFilterFunc: func(ctx context.Context, filter PositionFilter) ([]Position, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(positionRepo, &mockStockRepository{}, noopLogger{})
		_, err := service.Update(context.Background(), UpdatePositionInput{
			WalletID:     "wallet-id",
			PositionID:   "p1",
			Quantity:     20,
			AveragePrice: 2000,
		})

		assert.Error(t, err)
	})
}

func TestService_ConsolidateByWallet(t *testing.T) {
	t.Run("success - recalculates derivatives and portfolio percent for multiple positions", func(t *testing.T) {
		positionRepo := newMemoryPositionRepository()
		positionRepo.positions["p1"] = &Position{
			ID: "p1", WalletID: "wallet-id", StockID: "s1", Quantity: 100, AveragePrice: 1000,
		}
		positionRepo.positions["p2"] = &Position{
			ID: "p2", WalletID: "wallet-id", StockID: "s2", Quantity: 50, AveragePrice: 2000,
		}
		positionRepo.prices["s1"] = 1500
		positionRepo.prices["s2"] = 2000

		service := NewService(positionRepo, &mockStockRepository{}, noopLogger{})
		err := service.ConsolidateByWallet(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(100000), positionRepo.positions["p1"].Invested)
		assert.Equal(t, int64(150000), positionRepo.positions["p1"].Balance)
		assert.Equal(t, int64(50000), positionRepo.positions["p1"].VariationValue)
		assert.InDelta(t, 50.0, positionRepo.positions["p1"].VariationPercent, 0.001)
		assert.InDelta(t, 50.0, positionRepo.positions["p1"].PortfolioPercent, 0.001)
		assert.Equal(t, int64(100000), positionRepo.positions["p2"].Invested)
		assert.Equal(t, int64(100000), positionRepo.positions["p2"].Balance)
		assert.Equal(t, int64(0), positionRepo.positions["p2"].VariationValue)
		assert.InDelta(t, 0.0, positionRepo.positions["p2"].VariationPercent, 0.001)
		assert.InDelta(t, 50.0, positionRepo.positions["p2"].PortfolioPercent, 0.001)
	})

	t.Run("success - treats missing current price as zero", func(t *testing.T) {
		positionRepo := newMemoryPositionRepository()
		positionRepo.positions["p1"] = &Position{
			ID: "p1", WalletID: "wallet-id", StockID: "s1", Quantity: 10, AveragePrice: 1000,
		}

		service := NewService(positionRepo, &mockStockRepository{}, noopLogger{})
		err := service.ConsolidateByWallet(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(0), positionRepo.positions["p1"].CurrentPrice)
		assert.Equal(t, int64(0), positionRepo.positions["p1"].Balance)
		assert.Equal(t, int64(-10000), positionRepo.positions["p1"].VariationValue)
		assert.InDelta(t, -100.0, positionRepo.positions["p1"].VariationPercent, 0.001)
	})

	t.Run("error - returns error when update fails", func(t *testing.T) {
		positionRepo := &mockPositionRepository{
			runInTransactionFunc: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
			findByFilterFunc: func(ctx context.Context, filter PositionFilter) ([]Position, error) {
				return []Position{
					{ID: "p1", WalletID: filter.WalletID, StockID: "s1", Quantity: 10, AveragePrice: 1000},
				}, nil
			},
			findCurrentPricesMapFunc: func(ctx context.Context, stockIDs []string) (map[string]int64, error) {
				return map[string]int64{"s1": 1000}, nil
			},
			updateFunc: func(ctx context.Context, position *Position) error {
				return errors.New("update error")
			},
		}

		service := NewService(positionRepo, &mockStockRepository{}, noopLogger{})
		err := service.ConsolidateByWallet(context.Background(), "wallet-id")

		assert.Error(t, err)
	})
}
