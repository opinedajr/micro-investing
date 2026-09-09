package position

import (
	"context"
	"testing"

	"github.com/opinedajr/micro-investing/internal/infrastructure/database"
	"github.com/opinedajr/micro-investing/internal/stock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSQLiteRepository(t *testing.T) (*SQLiteRepository, context.Context) {
	gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
	require.NoError(t, err)
	require.NoError(t, gormDB.AutoMigrate(&Position{}, &stock.Stock{}, &stock.CurrentPrice{}))
	return NewSQLiteRepository(gormDB), context.Background()
}

func createStock(t *testing.T, repo *SQLiteRepository, ctx context.Context, id string, ticker string, rank int8) {
	s := &stock.Stock{ID: id, Ticker: ticker, Name: ticker, Sector: "Sector", Rank: rank}
	require.NoError(t, repo.txFromContext(ctx).Create(s).Error)
}

func TestSQLiteRepository_Create(t *testing.T) {
	t.Run("success - creates position with generated id", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		position := &Position{
			WalletID:     "wallet-id",
			StockID:      "stock-id",
			Quantity:     100,
			AveragePrice: 5000,
		}

		err := repo.Create(ctx, position)

		assert.NoError(t, err)
		assert.NotEmpty(t, position.ID)
	})

	t.Run("error - duplicate wallet/stock violates unique index", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		first := &Position{WalletID: "wallet-id", StockID: "stock-id", Quantity: 100, AveragePrice: 5000}
		second := &Position{WalletID: "wallet-id", StockID: "stock-id", Quantity: 50, AveragePrice: 6000}

		require.NoError(t, repo.Create(ctx, first))
		err := repo.Create(ctx, second)

		assert.Error(t, err)
	})
}

func TestSQLiteRepository_FindByID(t *testing.T) {
	t.Run("success - finds existing position", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		position := &Position{WalletID: "wallet-id", StockID: "stock-id", Quantity: 100, AveragePrice: 5000}
		require.NoError(t, repo.Create(ctx, position))

		found, err := repo.FindByID(ctx, position.ID)

		assert.NoError(t, err)
		assert.Equal(t, position.ID, found.ID)
	})

	t.Run("error - returns not found for missing id", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)

		_, err := repo.FindByID(ctx, "missing-id")

		assert.ErrorIs(t, err, ErrPositionNotFound)
	})
}

func TestSQLiteRepository_FindByWalletAndStockID(t *testing.T) {
	t.Run("success - finds existing position", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		position := &Position{WalletID: "wallet-id", StockID: "stock-id", Quantity: 100, AveragePrice: 5000}
		require.NoError(t, repo.Create(ctx, position))

		found, err := repo.FindByWalletAndStockID(ctx, "wallet-id", "stock-id")

		assert.NoError(t, err)
		assert.Equal(t, position.ID, found.ID)
	})

	t.Run("error - returns not found for missing pair", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)

		_, err := repo.FindByWalletAndStockID(ctx, "wallet-id", "stock-id")

		assert.ErrorIs(t, err, ErrPositionNotFound)
	})
}

func TestSQLiteRepository_FindByFilter(t *testing.T) {
	t.Run("success - returns positions for wallet ordered by balance desc by default", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		createStock(t, repo, ctx, "s1", "PETR4", 10)
		createStock(t, repo, ctx, "s2", "VALE3", 9)
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s1", Quantity: 10, AveragePrice: 1000, Balance: 5000, Invested: 10000}))
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s2", Quantity: 10, AveragePrice: 1000, Balance: 10000, Invested: 10000}))
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-b", StockID: "s3", Quantity: 10, AveragePrice: 1000, Balance: 20000, Invested: 10000}))

		positions, err := repo.FindByFilter(ctx, PositionFilter{WalletID: "wallet-a"})

		assert.NoError(t, err)
		require.Len(t, positions, 2)
		assert.Equal(t, "s2", positions[0].StockID)
		assert.Equal(t, "s1", positions[1].StockID)
	})

	t.Run("success - filters by ticker partial case-insensitive", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		createStock(t, repo, ctx, "s1", "PETR4", 10)
		createStock(t, repo, ctx, "s2", "VALE3", 9)
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s1", Quantity: 10, AveragePrice: 1000, Balance: 5000, Invested: 10000}))
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s2", Quantity: 10, AveragePrice: 1000, Balance: 10000, Invested: 10000}))

		positions, err := repo.FindByFilter(ctx, PositionFilter{WalletID: "wallet-a", Ticker: "pet"})

		assert.NoError(t, err)
		require.Len(t, positions, 1)
		assert.Equal(t, "s1", positions[0].StockID)
	})

	t.Run("success - returns empty slice when ticker filter has no match", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		createStock(t, repo, ctx, "s1", "PETR4", 10)
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s1", Quantity: 10, AveragePrice: 1000, Balance: 5000, Invested: 10000}))

		positions, err := repo.FindByFilter(ctx, PositionFilter{WalletID: "wallet-a", Ticker: "MISSING"})

		assert.NoError(t, err)
		assert.Empty(t, positions)
	})

	t.Run("success - sorts by ticker ascending", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		createStock(t, repo, ctx, "s1", "VALE3", 10)
		createStock(t, repo, ctx, "s2", "PETR4", 10)
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s1", Quantity: 10, AveragePrice: 1000, Balance: 10000, Invested: 10000}))
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s2", Quantity: 10, AveragePrice: 1000, Balance: 5000, Invested: 10000}))

		positions, err := repo.FindByFilter(ctx, PositionFilter{WalletID: "wallet-a", Sort: "ticker"})

		assert.NoError(t, err)
		require.Len(t, positions, 2)
		assert.Equal(t, "s2", positions[0].StockID)
		assert.Equal(t, "s1", positions[1].StockID)
	})

	t.Run("success - sorts by ticker descending", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		createStock(t, repo, ctx, "s1", "VALE3", 10)
		createStock(t, repo, ctx, "s2", "PETR4", 10)
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s1", Quantity: 10, AveragePrice: 1000, Balance: 10000, Invested: 10000}))
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s2", Quantity: 10, AveragePrice: 1000, Balance: 5000, Invested: 10000}))

		positions, err := repo.FindByFilter(ctx, PositionFilter{WalletID: "wallet-a", Sort: "-ticker"})

		assert.NoError(t, err)
		require.Len(t, positions, 2)
		assert.Equal(t, "s1", positions[0].StockID)
		assert.Equal(t, "s2", positions[1].StockID)
	})

	t.Run("success - sorts by rank ascending", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		createStock(t, repo, ctx, "s1", "A", 5)
		createStock(t, repo, ctx, "s2", "B", 3)
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s1", Quantity: 10, AveragePrice: 1000, Balance: 10000, Invested: 10000}))
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s2", Quantity: 10, AveragePrice: 1000, Balance: 10000, Invested: 10000}))

		positions, err := repo.FindByFilter(ctx, PositionFilter{WalletID: "wallet-a", Sort: "rank"})

		assert.NoError(t, err)
		require.Len(t, positions, 2)
		assert.Equal(t, "s2", positions[0].StockID)
		assert.Equal(t, "s1", positions[1].StockID)
	})

	t.Run("success - sorts by invested descending", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		createStock(t, repo, ctx, "s1", "A", 10)
		createStock(t, repo, ctx, "s2", "B", 10)
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s1", Quantity: 10, AveragePrice: 1000, Balance: 10000, Invested: 5000}))
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s2", Quantity: 10, AveragePrice: 1000, Balance: 10000, Invested: 15000}))

		positions, err := repo.FindByFilter(ctx, PositionFilter{WalletID: "wallet-a", Sort: "-invested"})

		assert.NoError(t, err)
		require.Len(t, positions, 2)
		assert.Equal(t, "s2", positions[0].StockID)
		assert.Equal(t, "s1", positions[1].StockID)
	})

	t.Run("success - sorts by variation_percent ascending", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		createStock(t, repo, ctx, "s1", "A", 10)
		createStock(t, repo, ctx, "s2", "B", 10)
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s1", Quantity: 10, AveragePrice: 1000, Balance: 10000, Invested: 10000, VariationPercent: 50}))
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s2", Quantity: 10, AveragePrice: 1000, Balance: 10000, Invested: 10000, VariationPercent: -10}))

		positions, err := repo.FindByFilter(ctx, PositionFilter{WalletID: "wallet-a", Sort: "variation_percent"})

		assert.NoError(t, err)
		require.Len(t, positions, 2)
		assert.Equal(t, "s2", positions[0].StockID)
		assert.Equal(t, "s1", positions[1].StockID)
	})

	t.Run("success - sorts by portfolio_percent descending", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		createStock(t, repo, ctx, "s1", "A", 10)
		createStock(t, repo, ctx, "s2", "B", 10)
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s1", Quantity: 10, AveragePrice: 1000, Balance: 10000, Invested: 10000, PortfolioPercent: 25}))
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s2", Quantity: 10, AveragePrice: 1000, Balance: 10000, Invested: 10000, PortfolioPercent: 75}))

		positions, err := repo.FindByFilter(ctx, PositionFilter{WalletID: "wallet-a", Sort: "-portfolio_percent"})

		assert.NoError(t, err)
		require.Len(t, positions, 2)
		assert.Equal(t, "s2", positions[0].StockID)
		assert.Equal(t, "s1", positions[1].StockID)
	})

	t.Run("success - falls back to default sort when sort is not in whitelist", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		createStock(t, repo, ctx, "s1", "A", 10)
		createStock(t, repo, ctx, "s2", "B", 10)
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s1", Quantity: 10, AveragePrice: 1000, Balance: 5000, Invested: 10000}))
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s2", Quantity: 10, AveragePrice: 1000, Balance: 10000, Invested: 10000}))

		positions, err := repo.FindByFilter(ctx, PositionFilter{WalletID: "wallet-a", Sort: "invalid_column"})

		assert.NoError(t, err)
		require.Len(t, positions, 2)
		assert.Equal(t, "s2", positions[0].StockID)
		assert.Equal(t, "s1", positions[1].StockID)
	})
}

func TestSQLiteRepository_FindCurrentPricesMap(t *testing.T) {
	t.Run("success - returns prices for requested stock ids", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		gormDB := repo.db
		require.NoError(t, gormDB.Exec("CREATE TABLE stocks_current_prices (stock_id TEXT PRIMARY KEY, price INTEGER NOT NULL, updated_at DATETIME NOT NULL)").Error)
		require.NoError(t, gormDB.Exec("INSERT INTO stocks_current_prices (stock_id, price, updated_at) VALUES (?, ?, datetime('now'))", "s1", 1500).Error)
		require.NoError(t, gormDB.Exec("INSERT INTO stocks_current_prices (stock_id, price, updated_at) VALUES (?, ?, datetime('now'))", "s2", 2500).Error)

		prices, err := repo.FindCurrentPricesMap(ctx, []string{"s1", "s2", "s3"})

		assert.NoError(t, err)
		assert.Equal(t, int64(1500), prices["s1"])
		assert.Equal(t, int64(2500), prices["s2"])
		assert.NotContains(t, prices, "s3")
	})
}

func TestSQLiteRepository_SumInvestedByWallet(t *testing.T) {
	t.Run("success - sums invested for wallet", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s1", Quantity: 10, AveragePrice: 1000, Invested: 10000}))
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-a", StockID: "s2", Quantity: 5, AveragePrice: 2000, Invested: 10000}))
		require.NoError(t, repo.Create(ctx, &Position{WalletID: "wallet-b", StockID: "s3", Quantity: 1, AveragePrice: 5000, Invested: 5000}))

		total, err := repo.SumInvestedByWallet(ctx, "wallet-a")

		assert.NoError(t, err)
		assert.Equal(t, int64(20000), total)
	})
}

func TestSQLiteRepository_Update(t *testing.T) {
	t.Run("success - updates position", func(t *testing.T) {
		repo, ctx := setupSQLiteRepository(t)
		position := &Position{WalletID: "wallet-id", StockID: "stock-id", Quantity: 10, AveragePrice: 1000}
		require.NoError(t, repo.Create(ctx, position))

		position.Balance = 5000
		position.Invested = 10000
		require.NoError(t, repo.Update(ctx, position))

		found, err := repo.FindByID(ctx, position.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(5000), found.Balance)
		assert.Equal(t, int64(10000), found.Invested)
	})
}
