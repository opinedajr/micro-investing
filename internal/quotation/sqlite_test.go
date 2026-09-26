package quotation

import (
	"context"
	"testing"
	"time"

	"github.com/opinedajr/micro-investing/internal/infrastructure/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type currentPriceRow struct {
	StockID   string
	Price     int64
	UpdatedAt time.Time
}

func setupSQLiteRepository(t *testing.T) (*SQLiteRepository, *gorm.DB, context.Context) {
	t.Helper()
	gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
	require.NoError(t, err)
	require.NoError(t, gormDB.Exec("CREATE TABLE stocks_current_prices (stock_id TEXT PRIMARY KEY, price INTEGER NOT NULL, updated_at DATETIME NOT NULL)").Error)
	return NewSQLiteRepository(gormDB), gormDB, context.Background()
}

func readPrices(t *testing.T, gormDB *gorm.DB) []currentPriceRow {
	t.Helper()
	var rows []currentPriceRow
	require.NoError(t, gormDB.Raw("SELECT stock_id, price, updated_at FROM stocks_current_prices ORDER BY stock_id").Scan(&rows).Error)
	return rows
}

func TestSQLiteRepository_UpsertCurrentPrices(t *testing.T) {
	t.Run("success - inserts new current prices", func(t *testing.T) {
		repo, gormDB, ctx := setupSQLiteRepository(t)

		err := repo.UpsertCurrentPrices(ctx, []CurrentPrice{
			{StockID: "s1", Price: 1500},
			{StockID: "s2", Price: 2500},
		})

		assert.NoError(t, err)
		rows := readPrices(t, gormDB)
		require.Len(t, rows, 2)
		assert.Equal(t, "s1", rows[0].StockID)
		assert.Equal(t, int64(1500), rows[0].Price)
		assert.False(t, rows[0].UpdatedAt.IsZero())
		assert.Equal(t, "s2", rows[1].StockID)
		assert.Equal(t, int64(2500), rows[1].Price)
	})

	t.Run("success - overwrites price and refreshes updated_at on conflict", func(t *testing.T) {
		repo, gormDB, ctx := setupSQLiteRepository(t)
		oldUpdatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		require.NoError(t, gormDB.Exec(
			"INSERT INTO stocks_current_prices (stock_id, price, updated_at) VALUES (?, ?, ?)",
			"s1", 1500, oldUpdatedAt,
		).Error)

		err := repo.UpsertCurrentPrices(ctx, []CurrentPrice{
			{StockID: "s1", Price: 2000, UpdatedAt: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)},
		})

		assert.NoError(t, err)
		rows := readPrices(t, gormDB)
		require.Len(t, rows, 1)
		assert.Equal(t, int64(2000), rows[0].Price)
		assert.True(t, rows[0].UpdatedAt.After(oldUpdatedAt), "updated_at deve ser sobrescrito com o momento da execução")
	})

	t.Run("success - upserts mixed batch of new and existing prices", func(t *testing.T) {
		repo, gormDB, ctx := setupSQLiteRepository(t)
		oldUpdatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		require.NoError(t, gormDB.Exec(
			"INSERT INTO stocks_current_prices (stock_id, price, updated_at) VALUES (?, ?, ?)",
			"s1", 1500, oldUpdatedAt,
		).Error)

		err := repo.UpsertCurrentPrices(ctx, []CurrentPrice{
			{StockID: "s1", Price: 1800},
			{StockID: "s2", Price: 3000},
		})

		assert.NoError(t, err)
		rows := readPrices(t, gormDB)
		require.Len(t, rows, 2)
		assert.Equal(t, "s1", rows[0].StockID)
		assert.Equal(t, int64(1800), rows[0].Price)
		assert.True(t, rows[0].UpdatedAt.After(oldUpdatedAt))
		assert.Equal(t, "s2", rows[1].StockID)
		assert.Equal(t, int64(3000), rows[1].Price)
	})

	t.Run("success - no-op for empty slice", func(t *testing.T) {
		repo, gormDB, ctx := setupSQLiteRepository(t)

		err := repo.UpsertCurrentPrices(ctx, []CurrentPrice{})

		assert.NoError(t, err)
		assert.Empty(t, readPrices(t, gormDB))
	})

	t.Run("error - rolls back entire transaction when an upsert fails mid-batch", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)
		require.NoError(t, gormDB.Exec("CREATE TABLE stocks_current_prices (stock_id TEXT PRIMARY KEY, price INTEGER NOT NULL CHECK(price > 0), updated_at DATETIME NOT NULL)").Error)
		repo := NewSQLiteRepository(gormDB)
		ctx := context.Background()
		oldUpdatedAt := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
		require.NoError(t, gormDB.Exec(
			"INSERT INTO stocks_current_prices (stock_id, price, updated_at) VALUES (?, ?, ?)",
			"s0", 500, oldUpdatedAt,
		).Error)

		err = repo.UpsertCurrentPrices(ctx, []CurrentPrice{
			{StockID: "s1", Price: 1500},
			{StockID: "s2", Price: -100},
			{StockID: "s3", Price: 2500},
		})

		assert.Error(t, err)
		rows := readPrices(t, gormDB)
		require.Len(t, rows, 1)
		assert.Equal(t, "s0", rows[0].StockID)
		assert.Equal(t, int64(500), rows[0].Price)
		assert.Equal(t, oldUpdatedAt, rows[0].UpdatedAt)
	})
}
