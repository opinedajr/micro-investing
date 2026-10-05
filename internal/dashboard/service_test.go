package dashboard

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/opinedajr/micro-investing/internal/dividends"
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

type mockDividendRepository struct {
	findByFilterFunc func(ctx context.Context, filter dividends.DividendFilter) ([]dividends.Dividend, error)
}

func (m *mockDividendRepository) Create(ctx context.Context, dividend *dividends.Dividend) error {
	return nil
}

func (m *mockDividendRepository) FindByFilter(ctx context.Context, filter dividends.DividendFilter) ([]dividends.Dividend, error) {
	if m.findByFilterFunc != nil {
		return m.findByFilterFunc(ctx, filter)
	}
	return nil, nil
}

func (m *mockDividendRepository) FindByWalletYear(ctx context.Context, walletID string, year int) (*dividends.Dividend, error) {
	return nil, nil
}

func (m *mockDividendRepository) FindByID(ctx context.Context, walletID string, id string) (*dividends.Dividend, error) {
	return nil, nil
}

func (m *mockDividendRepository) Update(ctx context.Context, dividend *dividends.Dividend) error {
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

		service := NewService(patrimonyRepo, positionRepo, &mockStockRepository{}, &mockDividendRepository{})
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

		service := NewService(patrimonyRepo, positionRepo, &mockStockRepository{}, &mockDividendRepository{})
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

		service := NewService(patrimonyRepo, positionRepo, &mockStockRepository{}, &mockDividendRepository{})
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

		service := NewService(patrimonyRepo, positionRepo, &mockStockRepository{}, &mockDividendRepository{})
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

		service := NewService(patrimonyRepo, positionRepo, &mockStockRepository{}, &mockDividendRepository{})
		_, err := service.Summary(context.Background(), "wallet-id")

		assert.Error(t, err)
	})

	t.Run("success - yearly dividends uses latest year not in the future", func(t *testing.T) {
		currentYear := time.Now().Year()
		dividendRepo := &mockDividendRepository{
			findByFilterFunc: func(ctx context.Context, filter dividends.DividendFilter) ([]dividends.Dividend, error) {
				return []dividends.Dividend{
					{WalletID: "wallet-id", Year: currentYear + 1, Amount: 999999},
					{WalletID: "wallet-id", Year: currentYear, Amount: 350000},
					{WalletID: "wallet-id", Year: currentYear - 2, Amount: 100000},
				}, nil
			},
		}

		service := NewService(&mockPatrimonyRepository{}, &mockPositionRepository{}, &mockStockRepository{}, dividendRepo)
		output, err := service.Summary(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(350000), output.YearlyDividends)
	})

	t.Run("success - yearly dividends falls back to previous year when current year missing", func(t *testing.T) {
		currentYear := time.Now().Year()
		dividendRepo := &mockDividendRepository{
			findByFilterFunc: func(ctx context.Context, filter dividends.DividendFilter) ([]dividends.Dividend, error) {
				return []dividends.Dividend{
					{WalletID: "wallet-id", Year: currentYear + 1, Amount: 999999},
					{WalletID: "wallet-id", Year: currentYear - 1, Amount: 250000},
				}, nil
			},
		}

		service := NewService(&mockPatrimonyRepository{}, &mockPositionRepository{}, &mockStockRepository{}, dividendRepo)
		output, err := service.Summary(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(250000), output.YearlyDividends)
	})

	t.Run("success - yearly dividends is zero when only future years exist", func(t *testing.T) {
		currentYear := time.Now().Year()
		dividendRepo := &mockDividendRepository{
			findByFilterFunc: func(ctx context.Context, filter dividends.DividendFilter) ([]dividends.Dividend, error) {
				return []dividends.Dividend{
					{WalletID: "wallet-id", Year: currentYear + 1, Amount: 500000},
				}, nil
			},
		}

		service := NewService(&mockPatrimonyRepository{}, &mockPositionRepository{}, &mockStockRepository{}, dividendRepo)
		output, err := service.Summary(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(0), output.YearlyDividends)
	})

	t.Run("error - returns error when finding dividends fails", func(t *testing.T) {
		dividendRepo := &mockDividendRepository{
			findByFilterFunc: func(ctx context.Context, filter dividends.DividendFilter) ([]dividends.Dividend, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(&mockPatrimonyRepository{}, &mockPositionRepository{}, &mockStockRepository{}, dividendRepo)
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

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
		output, err := service.Allocation(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(1500000), output.Total)
		assert.Len(t, output.Items, 4)

		expectedPercentages := map[string]int{
			"stocks":            33,
			"fixed_income":      33,
			"emergency_reserve": 17,
			"liquid_cash":       17,
		}
		for _, item := range output.Items {
			assert.Equal(t, expectedPercentages[item.Type], item.Percentage)
		}
	})

	t.Run("success - rounds percentages to integer rounding halves away from zero", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			findLatestMonthByWalletFunc: func(ctx context.Context, walletID string) (int, int, error) {
				return 2026, 3, nil
			},
			sumByWalletYearMonthFunc: func(ctx context.Context, walletID string, year int, month int) ([]patrimony.TypeAmount, error) {
				return []patrimony.TypeAmount{
					{Type: patrimony.TypeStocks, Amount: 200000},
					{Type: patrimony.TypeFixedIncome, Amount: 150000},
					{Type: patrimony.TypeEmergencyReserve, Amount: 50000},
				}, nil
			},
		}

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
		output, err := service.Allocation(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(400000), output.Total)

		expectedPercentages := map[string]int{
			"stocks":            50,
			"fixed_income":      38,
			"emergency_reserve": 13,
		}
		for _, item := range output.Items {
			assert.Equal(t, expectedPercentages[item.Type], item.Percentage)
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

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
		output, err := service.Allocation(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(100000), output.Total)
		assert.Len(t, output.Items, 1)
		assert.Equal(t, "stocks", output.Items[0].Type)
		assert.Equal(t, int64(100000), output.Items[0].Amount)
		assert.Equal(t, 100, output.Items[0].Percentage)
	})

	t.Run("success - returns empty when wallet has no patrimony records", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			findLatestMonthByWalletFunc: func(ctx context.Context, walletID string) (int, int, error) {
				return 0, 0, nil
			},
		}

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
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

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
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

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
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

		service := NewService(&mockPatrimonyRepository{}, positionRepo, stockRepo, &mockDividendRepository{})
		output, err := service.Risk(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(500000), output.Total)
		assert.Len(t, output.Items, 2)

		expected := map[int8]int{
			3: 60,
			4: 40,
		}
		for _, item := range output.Items {
			assert.Equal(t, expected[item.Rank], item.Percentage)
		}
	})

	t.Run("success - rounds percentages to integer rounding halves away from zero", func(t *testing.T) {
		positionRepo := &mockPositionRepository{
			findByFilterFunc: func(ctx context.Context, filter position.PositionFilter) ([]position.Position, error) {
				return []position.Position{
					{StockID: "stock-1", Invested: 200000},
					{StockID: "stock-2", Invested: 150000},
					{StockID: "stock-3", Invested: 50000},
				}, nil
			},
		}
		stockRepo := &mockStockRepository{
			findByIDsFunc: func(ctx context.Context, ids []string) ([]stock.Stock, error) {
				return []stock.Stock{
					{ID: "stock-1", Rank: 5},
					{ID: "stock-2", Rank: 4},
					{ID: "stock-3", Rank: 3},
				}, nil
			},
		}

		service := NewService(&mockPatrimonyRepository{}, positionRepo, stockRepo, &mockDividendRepository{})
		output, err := service.Risk(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Equal(t, int64(400000), output.Total)

		expected := map[int8]int{
			5: 50,
			4: 38,
			3: 13,
		}
		for _, item := range output.Items {
			assert.Equal(t, expected[item.Rank], item.Percentage)
		}
	})

	t.Run("success - returns empty when wallet has no positions", func(t *testing.T) {
		positionRepo := &mockPositionRepository{
			findByFilterFunc: func(ctx context.Context, filter position.PositionFilter) ([]position.Position, error) {
				return []position.Position{}, nil
			},
		}
		stockRepo := &mockStockRepository{}

		service := NewService(&mockPatrimonyRepository{}, positionRepo, stockRepo, &mockDividendRepository{})
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

		service := NewService(&mockPatrimonyRepository{}, positionRepo, stockRepo, &mockDividendRepository{})
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

		service := NewService(&mockPatrimonyRepository{}, positionRepo, stockRepo, &mockDividendRepository{})
		_, err := service.Risk(context.Background(), "wallet-id")

		assert.Error(t, err)
	})
}

func TestResolveEvolutionPeriod(t *testing.T) {
	t.Run("default - last 12 months from reference", func(t *testing.T) {
		reference := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
		input := EvolutionInput{}

		startYear, startMonth, endYear, endMonth, err := resolveEvolutionPeriod(reference, input)

		assert.NoError(t, err)
		assert.Equal(t, 2025, startYear)
		assert.Equal(t, 8, startMonth)
		assert.Equal(t, 2026, endYear)
		assert.Equal(t, 7, endMonth)
	})

	t.Run("year filter - full year", func(t *testing.T) {
		reference := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
		input := EvolutionInput{Year: 2025}

		startYear, startMonth, endYear, endMonth, err := resolveEvolutionPeriod(reference, input)

		assert.NoError(t, err)
		assert.Equal(t, 2025, startYear)
		assert.Equal(t, 1, startMonth)
		assert.Equal(t, 2025, endYear)
		assert.Equal(t, 12, endMonth)
	})

	t.Run("year and quarter filter - Q2", func(t *testing.T) {
		reference := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
		input := EvolutionInput{Year: 2026, Quarter: 2}

		startYear, startMonth, endYear, endMonth, err := resolveEvolutionPeriod(reference, input)

		assert.NoError(t, err)
		assert.Equal(t, 2026, startYear)
		assert.Equal(t, 4, startMonth)
		assert.Equal(t, 2026, endYear)
		assert.Equal(t, 6, endMonth)
	})

	t.Run("year and quarter filter - Q4", func(t *testing.T) {
		reference := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
		input := EvolutionInput{Year: 2025, Quarter: 4}

		startYear, startMonth, endYear, endMonth, err := resolveEvolutionPeriod(reference, input)

		assert.NoError(t, err)
		assert.Equal(t, 2025, startYear)
		assert.Equal(t, 10, startMonth)
		assert.Equal(t, 2025, endYear)
		assert.Equal(t, 12, endMonth)
	})

	t.Run("error - quarter greater than 4", func(t *testing.T) {
		reference := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
		input := EvolutionInput{Year: 2026, Quarter: 5}

		_, _, _, _, err := resolveEvolutionPeriod(reference, input)

		assert.Error(t, err)
	})

	t.Run("year filter - quarter zero means full year", func(t *testing.T) {
		reference := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
		input := EvolutionInput{Year: 2026, Quarter: 0}

		startYear, startMonth, endYear, endMonth, err := resolveEvolutionPeriod(reference, input)

		assert.NoError(t, err)
		assert.Equal(t, 2026, startYear)
		assert.Equal(t, 1, startMonth)
		assert.Equal(t, 2026, endYear)
		assert.Equal(t, 12, endMonth)
	})

	t.Run("error - negative quarter", func(t *testing.T) {
		reference := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
		input := EvolutionInput{Year: 2026, Quarter: -1}

		_, _, _, _, err := resolveEvolutionPeriod(reference, input)

		assert.Error(t, err)
	})

	t.Run("error - quarter without year", func(t *testing.T) {
		reference := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
		input := EvolutionInput{Quarter: 1}

		_, _, _, _, err := resolveEvolutionPeriod(reference, input)

		assert.Error(t, err)
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

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
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

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
		output, err := service.Evolution(context.Background(), "wallet-id", EvolutionInput{Year: 2026, Quarter: 2})

		assert.NoError(t, err)
		assert.Len(t, output.Total, 3)
		assert.Equal(t, int64(900000), output.Total[0].Amount)
		assert.Equal(t, int64(900000), output.Total[1].Amount)
		assert.Equal(t, int64(1150000), output.Total[2].Amount)

		assert.Equal(t, int64(500000), output.ByCategory.FixedIncome[1].Amount)
		assert.Equal(t, int64(400000), output.ByCategory.Stocks[1].Amount)
	})

	t.Run("success - carry forward per category when partial data", func(t *testing.T) {
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
					},
					"2026-6": {
						{Type: patrimony.TypeFixedIncome, Amount: 550000},
					},
				}
				key := fmt.Sprintf("%d-%d", year, month)
				return data[key], nil
			},
		}

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
		output, err := service.Evolution(context.Background(), "wallet-id", EvolutionInput{Year: 2026, Quarter: 2})

		assert.NoError(t, err)
		assert.Len(t, output.Total, 3)
		assert.Equal(t, int64(1200000), output.Total[0].Amount)
		assert.Equal(t, int64(1000000), output.Total[1].Amount)
		assert.Equal(t, int64(550000), output.Total[2].Amount)

		assert.Len(t, output.ByCategory.FixedIncome, 3)
		assert.Equal(t, int64(500000), output.ByCategory.FixedIncome[0].Amount)
		assert.Equal(t, int64(500000), output.ByCategory.FixedIncome[1].Amount)
		assert.Equal(t, int64(550000), output.ByCategory.FixedIncome[2].Amount)

		assert.Len(t, output.ByCategory.Stocks, 3)
		assert.Equal(t, int64(400000), output.ByCategory.Stocks[0].Amount)
		assert.Equal(t, int64(500000), output.ByCategory.Stocks[1].Amount)
		assert.Equal(t, int64(500000), output.ByCategory.Stocks[2].Amount)

		assert.Len(t, output.ByCategory.EmergencyReserve, 3)
		assert.Equal(t, int64(300000), output.ByCategory.EmergencyReserve[0].Amount)
		assert.Equal(t, int64(300000), output.ByCategory.EmergencyReserve[1].Amount)
		assert.Equal(t, int64(300000), output.ByCategory.EmergencyReserve[2].Amount)
	})

	t.Run("error - returns error for invalid quarter", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{}

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
		_, err := service.Evolution(context.Background(), "wallet-id", EvolutionInput{Year: 2026, Quarter: 5})

		assert.Error(t, err)
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

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
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

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
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

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
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

		service := NewService(patrimonyRepo, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
		_, err := service.Evolution(context.Background(), "wallet-id", EvolutionInput{Year: 2026})

		assert.Error(t, err)
	})
}

func TestService_Dividends(t *testing.T) {
	t.Run("success - returns all years ordered ascending without id", func(t *testing.T) {
		dividendRepo := &mockDividendRepository{
			findByFilterFunc: func(ctx context.Context, filter dividends.DividendFilter) ([]dividends.Dividend, error) {
				assert.Equal(t, "wallet-id", filter.WalletID)
				return []dividends.Dividend{
					{ID: "div-3", WalletID: "wallet-id", Year: 2026, Amount: 350000},
					{ID: "div-2", WalletID: "wallet-id", Year: 2024, Amount: 150000},
					{ID: "div-1", WalletID: "wallet-id", Year: 2023, Amount: 100000},
				}, nil
			},
		}

		service := NewService(&mockPatrimonyRepository{}, &mockPositionRepository{}, &mockStockRepository{}, dividendRepo)
		output, err := service.Dividends(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.Len(t, output.Items, 3)
		assert.Equal(t, 2023, output.Items[0].Year)
		assert.Equal(t, int64(100000), output.Items[0].Amount)
		assert.Equal(t, 2024, output.Items[1].Year)
		assert.Equal(t, int64(150000), output.Items[1].Amount)
		assert.Equal(t, 2026, output.Items[2].Year)
		assert.Equal(t, int64(350000), output.Items[2].Amount)
	})

	t.Run("success - returns empty items when wallet has no dividends", func(t *testing.T) {
		service := NewService(&mockPatrimonyRepository{}, &mockPositionRepository{}, &mockStockRepository{}, &mockDividendRepository{})
		output, err := service.Dividends(context.Background(), "wallet-id")

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.NotNil(t, output.Items)
		assert.Len(t, output.Items, 0)
	})

	t.Run("error - returns error when finding dividends fails", func(t *testing.T) {
		dividendRepo := &mockDividendRepository{
			findByFilterFunc: func(ctx context.Context, filter dividends.DividendFilter) ([]dividends.Dividend, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(&mockPatrimonyRepository{}, &mockPositionRepository{}, &mockStockRepository{}, dividendRepo)
		_, err := service.Dividends(context.Background(), "wallet-id")

		assert.Error(t, err)
	})
}
