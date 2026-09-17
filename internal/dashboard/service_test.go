package dashboard

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

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
	findByFilterFunc        func(ctx context.Context, filter position.PositionFilter) ([]position.Position, error)
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
	if m.findByFilterFunc != nil {
		return m.findByFilterFunc(ctx, filter)
	}
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

type mockStockRepository struct {
	findByIDsFunc func(ctx context.Context, ids []string) ([]stock.Stock, error)
}

func (m *mockStockRepository) Create(ctx context.Context, s *stock.Stock) error {
	return nil
}

func (m *mockStockRepository) FindByTicker(ctx context.Context, ticker string) (*stock.Stock, error) {
	return nil, nil
}

func (m *mockStockRepository) FindByID(ctx context.Context, id string) (*stock.Stock, error) {
	return nil, nil
}

func (m *mockStockRepository) FindByIDs(ctx context.Context, ids []string) ([]stock.Stock, error) {
	if m.findByIDsFunc != nil {
		return m.findByIDsFunc(ctx, ids)
	}
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

func TestService_Allocation(t *testing.T) {
	t.Run("success - returns aggregated allocation with percentages", func(t *testing.T) {
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

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{})
		output, err := service.Allocation(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(1500000), output.Total)
		assert.Len(t, output.Items, 4)

		expectedPercentages := map[string]float64{
			"stocks":            33.33,
			"fixed_income":      33.33,
			"emergency_reserve": 16.67,
			"liquid_cash":       16.67,
		}
		for _, item := range output.Items {
			assert.InDelta(t, expectedPercentages[item.Type], item.Percentage, 0.01)
		}
	})

	t.Run("success - omits categories without balance", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			findLatestMonthByWalletFunc: func(ctx context.Context, walletID string) (int, int, error) {
				return 2026, 7, nil
			},
			sumByWalletYearMonthFunc: func(ctx context.Context, walletID string, year int, month int) ([]patrimony.TypeAmount, error) {
				return []patrimony.TypeAmount{
					{Type: patrimony.TypeStocks, Amount: 100000},
					{Type: patrimony.TypeFixedIncome, Amount: 0},
					{Type: patrimony.TypeFIIs, Amount: 0},
				}, nil
			},
		}

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{})
		output, err := service.Allocation(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(100000), output.Total)
		assert.Len(t, output.Items, 1)
		assert.Equal(t, "stocks", output.Items[0].Type)
		assert.Equal(t, int64(100000), output.Items[0].Amount)
		assert.InDelta(t, 100.0, output.Items[0].Percentage, 0.01)
	})

	t.Run("success - returns empty when wallet has no patrimony records", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			findLatestMonthByWalletFunc: func(ctx context.Context, walletID string) (int, int, error) {
				return 0, 0, nil
			},
		}

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{})
		output, err := service.Allocation(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(0), output.Total)
		assert.Empty(t, output.Items)
	})

	t.Run("error - returns error when finding latest month fails", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			findLatestMonthByWalletFunc: func(ctx context.Context, walletID string) (int, int, error) {
				return 0, 0, errors.New("database error")
			},
		}

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{})
		_, err := service.Allocation(context.Background(), "wallet-id")

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

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{})
		_, err := service.Allocation(context.Background(), "wallet-id")

		assert.Error(t, err)
	})
}

func TestService_Risk(t *testing.T) {
	t.Run("success - returns aggregated risk by stock rank", func(t *testing.T) {
		positionRepo := &mockPositionRepository{
			findByFilterFunc: func(ctx context.Context, filter position.PositionFilter) ([]position.Position, error) {
				return []position.Position{
					{StockID: "stock-1", Invested: 300000},
					{StockID: "stock-2", Invested: 200000},
				}, nil
			},
		}
		stockRepo := &mockStockRepository{
			findByIDsFunc: func(ctx context.Context, ids []string) ([]stock.Stock, error) {
				return []stock.Stock{
					{ID: "stock-1", Rank: 3},
					{ID: "stock-2", Rank: 4},
				}, nil
			},
		}

		service := NewService(&mockPatrimonyRepository{}, positionRepo, stockRepo)
		output, err := service.Risk(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(500000), output.Total)
		assert.Len(t, output.Items, 2)

		expected := map[int8]float64{
			3: 60.0,
			4: 40.0,
		}
		for _, item := range output.Items {
			assert.InDelta(t, expected[item.Rank], item.Percentage, 0.01)
		}
	})

	t.Run("success - returns empty when wallet has no positions", func(t *testing.T) {
		positionRepo := &mockPositionRepository{
			findByFilterFunc: func(ctx context.Context, filter position.PositionFilter) ([]position.Position, error) {
				return []position.Position{}, nil
			},
		}
		stockRepo := &mockStockRepository{}

		service := NewService(&mockPatrimonyRepository{}, positionRepo, stockRepo)
		output, err := service.Risk(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(0), output.Total)
		assert.Empty(t, output.Items)
	})

	t.Run("error - returns error when finding positions fails", func(t *testing.T) {
		positionRepo := &mockPositionRepository{
			findByFilterFunc: func(ctx context.Context, filter position.PositionFilter) ([]position.Position, error) {
				return nil, errors.New("database error")
			},
		}
		stockRepo := &mockStockRepository{}

		service := NewService(&mockPatrimonyRepository{}, positionRepo, stockRepo)
		_, err := service.Risk(context.Background(), "wallet-id")

		assert.Error(t, err)
	})

	t.Run("error - returns error when finding stocks fails", func(t *testing.T) {
		positionRepo := &mockPositionRepository{
			findByFilterFunc: func(ctx context.Context, filter position.PositionFilter) ([]position.Position, error) {
				return []position.Position{
					{StockID: "stock-1", Invested: 300000},
				}, nil
			},
		}
		stockRepo := &mockStockRepository{
			findByIDsFunc: func(ctx context.Context, ids []string) ([]stock.Stock, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(&mockPatrimonyRepository{}, positionRepo, stockRepo)
		_, err := service.Risk(context.Background(), "wallet-id")

		assert.Error(t, err)
	})
}

func TestResolveEvolutionPeriod(t *testing.T) {
	t.Run("default - last 12 months from reference", func(t *testing.T) {
		reference := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
		input := EvolutionInput{}

		startYear, startMonth, endYear, endMonth := resolveEvolutionPeriod(reference, input)

		assert.Equal(t, 2025, startYear)
		assert.Equal(t, 8, startMonth)
		assert.Equal(t, 2026, endYear)
		assert.Equal(t, 7, endMonth)
	})

	t.Run("year filter - full year", func(t *testing.T) {
		reference := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
		input := EvolutionInput{Year: 2025}

		startYear, startMonth, endYear, endMonth := resolveEvolutionPeriod(reference, input)

		assert.Equal(t, 2025, startYear)
		assert.Equal(t, 1, startMonth)
		assert.Equal(t, 2025, endYear)
		assert.Equal(t, 12, endMonth)
	})

	t.Run("year and quarter filter - Q2", func(t *testing.T) {
		reference := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
		input := EvolutionInput{Year: 2026, Quarter: 2}

		startYear, startMonth, endYear, endMonth := resolveEvolutionPeriod(reference, input)

		assert.Equal(t, 2026, startYear)
		assert.Equal(t, 4, startMonth)
		assert.Equal(t, 2026, endYear)
		assert.Equal(t, 6, endMonth)
	})

	t.Run("year and quarter filter - Q4", func(t *testing.T) {
		reference := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
		input := EvolutionInput{Year: 2025, Quarter: 4}

		startYear, startMonth, endYear, endMonth := resolveEvolutionPeriod(reference, input)

		assert.Equal(t, 2025, startYear)
		assert.Equal(t, 10, startMonth)
		assert.Equal(t, 2025, endYear)
		assert.Equal(t, 12, endMonth)
	})
}

func TestService_Evolution(t *testing.T) {
	t.Run("success - filters by year and quarter", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			sumByWalletYearMonthFunc: func(ctx context.Context, walletID string, year int, month int) ([]patrimony.TypeAmount, error) {
				data := map[string][]patrimony.TypeAmount{
					"2026-4": {
						{Type: patrimony.TypeStocks, Amount: 400000},
						{Type: patrimony.TypeFixedIncome, Amount: 500000},
						{Type: patrimony.TypeEmergencyReserve, Amount: 300000},
					},
					"2026-5": {
						{Type: patrimony.TypeStocks, Amount: 500000},
						{Type: patrimony.TypeFixedIncome, Amount: 500000},
						{Type: patrimony.TypeEmergencyReserve, Amount: 350000},
					},
					"2026-6": {
						{Type: patrimony.TypeStocks, Amount: 600000},
						{Type: patrimony.TypeFixedIncome, Amount: 550000},
						{Type: patrimony.TypeEmergencyReserve, Amount: 350000},
					},
				}
				key := fmt.Sprintf("%d-%d", year, month)
				return data[key], nil
			},
		}

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{})
		output, err := service.Evolution(context.Background(), "wallet-id", EvolutionInput{Year: 2026, Quarter: 2})

		assert.NoError(t, err)
		assert.Len(t, output.Total, 3)
		assert.Equal(t, int64(1200000), output.Total[0].Amount)
		assert.Equal(t, int64(1350000), output.Total[1].Amount)
		assert.Equal(t, int64(1500000), output.Total[2].Amount)

		assert.Len(t, output.ByCategory.FixedIncome, 3)
		assert.Equal(t, int64(500000), output.ByCategory.FixedIncome[0].Amount)
		assert.Equal(t, int64(500000), output.ByCategory.FixedIncome[1].Amount)
		assert.Equal(t, int64(550000), output.ByCategory.FixedIncome[2].Amount)

		assert.Len(t, output.ByCategory.Stocks, 3)
		assert.Equal(t, int64(400000), output.ByCategory.Stocks[0].Amount)
		assert.Equal(t, int64(500000), output.ByCategory.Stocks[1].Amount)
		assert.Equal(t, int64(600000), output.ByCategory.Stocks[2].Amount)

		assert.Len(t, output.ByCategory.EmergencyReserve, 3)
		assert.Equal(t, int64(300000), output.ByCategory.EmergencyReserve[0].Amount)
		assert.Equal(t, int64(350000), output.ByCategory.EmergencyReserve[1].Amount)
		assert.Equal(t, int64(350000), output.ByCategory.EmergencyReserve[2].Amount)
	})

	t.Run("success - carry forward fills months without data", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			sumByWalletYearMonthFunc: func(ctx context.Context, walletID string, year int, month int) ([]patrimony.TypeAmount, error) {
				data := map[string][]patrimony.TypeAmount{
					"2026-4": {
						{Type: patrimony.TypeStocks, Amount: 400000},
						{Type: patrimony.TypeFixedIncome, Amount: 500000},
					},
					"2026-6": {
						{Type: patrimony.TypeStocks, Amount: 600000},
						{Type: patrimony.TypeFixedIncome, Amount: 550000},
					},
				}
				key := fmt.Sprintf("%d-%d", year, month)
				return data[key], nil
			},
		}

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{})
		output, err := service.Evolution(context.Background(), "wallet-id", EvolutionInput{Year: 2026, Quarter: 2})

		assert.NoError(t, err)
		assert.Len(t, output.Total, 3)
		assert.Equal(t, int64(900000), output.Total[0].Amount)
		assert.Equal(t, int64(900000), output.Total[1].Amount)
		assert.Equal(t, int64(1150000), output.Total[2].Amount)

		assert.Equal(t, int64(500000), output.ByCategory.FixedIncome[1].Amount)
		assert.Equal(t, int64(400000), output.ByCategory.Stocks[1].Amount)
	})

	t.Run("success - total includes all categories including fiis and liquid_cash", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			sumByWalletYearMonthFunc: func(ctx context.Context, walletID string, year int, month int) ([]patrimony.TypeAmount, error) {
				if year == 2026 && month == 4 {
					return []patrimony.TypeAmount{
						{Type: patrimony.TypeStocks, Amount: 400000},
						{Type: patrimony.TypeFIIs, Amount: 100000},
						{Type: patrimony.TypeFixedIncome, Amount: 500000},
						{Type: patrimony.TypeEmergencyReserve, Amount: 300000},
						{Type: patrimony.TypeLiquidCash, Amount: 50000},
					}, nil
				}
				return nil, nil
			},
		}

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{})
		output, err := service.Evolution(context.Background(), "wallet-id", EvolutionInput{Year: 2026, Quarter: 2})

		assert.NoError(t, err)
		assert.Len(t, output.Total, 3)
		assert.Equal(t, int64(1350000), output.Total[0].Amount)
		assert.Equal(t, int64(1350000), output.Total[1].Amount)
		assert.Equal(t, int64(1350000), output.Total[2].Amount)
		assert.Len(t, output.ByCategory.FixedIncome, 3)
		assert.Len(t, output.ByCategory.Stocks, 3)
		assert.Len(t, output.ByCategory.EmergencyReserve, 3)
	})

	t.Run("success - default last 12 months when no filters", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			sumByWalletYearMonthFunc: func(ctx context.Context, walletID string, year int, month int) ([]patrimony.TypeAmount, error) {
				return []patrimony.TypeAmount{{Type: patrimony.TypeFixedIncome, Amount: 100000}}, nil
			},
		}

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{})
		output, err := service.Evolution(context.Background(), "wallet-id", EvolutionInput{})

		assert.NoError(t, err)
		assert.Len(t, output.Total, 12)
		assert.Len(t, output.ByCategory.FixedIncome, 12)
	})

	t.Run("success - returns empty series when no data and no previous value to carry forward", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			sumByWalletYearMonthFunc: func(ctx context.Context, walletID string, year int, month int) ([]patrimony.TypeAmount, error) {
				return []patrimony.TypeAmount{}, nil
			},
		}

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{})
		output, err := service.Evolution(context.Background(), "wallet-id", EvolutionInput{Year: 2026, Quarter: 2})

		assert.NoError(t, err)
		assert.Len(t, output.Total, 3)
		assert.Equal(t, int64(0), output.Total[0].Amount)
		assert.Equal(t, int64(0), output.Total[1].Amount)
		assert.Equal(t, int64(0), output.Total[2].Amount)
	})

	t.Run("error - returns error when summing patrimony fails", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			sumByWalletYearMonthFunc: func(ctx context.Context, walletID string, year int, month int) ([]patrimony.TypeAmount, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{})
		_, err := service.Evolution(context.Background(), "wallet-id", EvolutionInput{Year: 2026})

		assert.Error(t, err)
	})
}
