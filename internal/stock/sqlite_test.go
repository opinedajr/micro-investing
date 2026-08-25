package stock

import (
	"context"
	"testing"

	"github.com/opinedajr/micro-investing/internal/infrastructure/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteRepository_Create(t *testing.T) {
	t.Run("success - creates stock with generated id", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		require.NoError(t, gormDB.AutoMigrate(&Stock{}))

		repo := NewSQLiteRepository(gormDB)
		stock := &Stock{Ticker: "PETR4", Name: "Petrobras PN", Sector: "Petróleo", Rank: 10}

		err = repo.Create(context.Background(), stock)
		assert.NoError(t, err)
		assert.NotEmpty(t, stock.ID)

		found, err := repo.FindByTicker(context.Background(), "PETR4")
		assert.NoError(t, err)
		assert.Equal(t, stock.ID, found.ID)
	})

	t.Run("error - duplicate ticker violates unique index", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		require.NoError(t, gormDB.AutoMigrate(&Stock{}))

		repo := NewSQLiteRepository(gormDB)
		first := &Stock{Ticker: "PETR4", Name: "Petrobras PN", Sector: "Petróleo", Rank: 10}
		second := &Stock{Ticker: "PETR4", Name: "Petrobras Copy", Sector: "Petróleo", Rank: 9}

		require.NoError(t, repo.Create(context.Background(), first))
		err = repo.Create(context.Background(), second)

		assert.Error(t, err)
	})
}

func TestSQLiteRepository_FindByTicker(t *testing.T) {
	t.Run("success - finds existing stock", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		require.NoError(t, gormDB.AutoMigrate(&Stock{}))

		repo := NewSQLiteRepository(gormDB)
		stock := &Stock{Ticker: "VALE3", Name: "Vale ON", Sector: "Mineração", Rank: 10}
		require.NoError(t, repo.Create(context.Background(), stock))

		found, err := repo.FindByTicker(context.Background(), "VALE3")
		assert.NoError(t, err)
		assert.Equal(t, "VALE3", found.Ticker)
	})

	t.Run("error - returns not found for missing ticker", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		require.NoError(t, gormDB.AutoMigrate(&Stock{}))

		repo := NewSQLiteRepository(gormDB)
		_, err = repo.FindByTicker(context.Background(), "MISS3")

		assert.ErrorIs(t, err, ErrStockNotFound)
	})
}

func TestSQLiteRepository_FindByID(t *testing.T) {
	t.Run("success - finds stock by id", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		require.NoError(t, gormDB.AutoMigrate(&Stock{}))

		repo := NewSQLiteRepository(gormDB)
		stock := &Stock{Ticker: "WEGE3", Name: "WEG ON", Sector: "Máquinas", Rank: 10}
		require.NoError(t, repo.Create(context.Background(), stock))

		found, err := repo.FindByID(context.Background(), stock.ID)
		assert.NoError(t, err)
		assert.Equal(t, stock.ID, found.ID)
		assert.Equal(t, "WEGE3", found.Ticker)
	})

	t.Run("error - returns not found for missing id", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		require.NoError(t, gormDB.AutoMigrate(&Stock{}))

		repo := NewSQLiteRepository(gormDB)
		_, err = repo.FindByID(context.Background(), "missing-id")

		assert.ErrorIs(t, err, ErrStockNotFound)
	})
}

func TestSQLiteRepository_List(t *testing.T) {
	t.Run("success - returns stocks ordered by ticker", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		require.NoError(t, gormDB.AutoMigrate(&Stock{}))

		repo := NewSQLiteRepository(gormDB)
		require.NoError(t, repo.Create(context.Background(), &Stock{Ticker: "VALE3", Name: "Vale", Sector: "Mineração", Rank: 10}))
		require.NoError(t, repo.Create(context.Background(), &Stock{Ticker: "PETR4", Name: "Petrobras", Sector: "Petróleo", Rank: 10}))

		stocks, err := repo.List(context.Background())
		assert.NoError(t, err)
		require.Len(t, stocks, 2)
		assert.Equal(t, "PETR4", stocks[0].Ticker)
		assert.Equal(t, "VALE3", stocks[1].Ticker)
	})
}

func TestSQLiteRepository_Seed(t *testing.T) {
	t.Run("success - seed is idempotent", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		require.NoError(t, gormDB.AutoMigrate(&Stock{}))

		repo := NewSQLiteRepository(gormDB)
		stocks := []Stock{
			{Ticker: "PETR4", Name: "Petrobras PN", Sector: "Petróleo", Rank: 10},
			{Ticker: "VALE3", Name: "Vale ON", Sector: "Mineração", Rank: 10},
		}

		require.NoError(t, repo.Seed(context.Background(), stocks, false))
		require.NoError(t, repo.Seed(context.Background(), stocks, false))

		result, err := repo.List(context.Background())
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("success - force updates existing stocks", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		require.NoError(t, gormDB.AutoMigrate(&Stock{}))

		repo := NewSQLiteRepository(gormDB)
		first := []Stock{{Ticker: "PETR4", Name: "Petrobras PN", Sector: "Petróleo", Rank: 10}}
		second := []Stock{{Ticker: "PETR4", Name: "Petrobras Updated", Sector: "Petróleo e Gás", Rank: 9}}

		require.NoError(t, repo.Seed(context.Background(), first, false))
		require.NoError(t, repo.Seed(context.Background(), second, true))

		found, err := repo.FindByTicker(context.Background(), "PETR4")
		require.NoError(t, err)
		assert.Equal(t, "Petrobras Updated", found.Name)
		assert.Equal(t, "Petróleo e Gás", found.Sector)
		assert.Equal(t, int8(9), found.Rank)
	})

	t.Run("success - manual edits preserved without force", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		require.NoError(t, gormDB.AutoMigrate(&Stock{}))

		repo := NewSQLiteRepository(gormDB)
		first := []Stock{{Ticker: "PETR4", Name: "Petrobras PN", Sector: "Petróleo", Rank: 10}}
		second := []Stock{{Ticker: "PETR4", Name: "Seed Name", Sector: "Petróleo", Rank: 10}}

		require.NoError(t, repo.Seed(context.Background(), first, false))

		found, err := repo.FindByTicker(context.Background(), "PETR4")
		require.NoError(t, err)
		found.Name = "Manual Edit"
		require.NoError(t, gormDB.Save(found).Error)

		require.NoError(t, repo.Seed(context.Background(), second, false))

		found, err = repo.FindByTicker(context.Background(), "PETR4")
		require.NoError(t, err)
		assert.Equal(t, "Manual Edit", found.Name)
	})
}
