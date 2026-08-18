package stock

import (
	"context"
	"testing"

	"github.com/opinedajr/micro-investing/internal/infrastructure/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteRepository_List(t *testing.T) {
	t.Run("success - lists all stocks", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		err = gormDB.AutoMigrate(&Stock{})
		require.NoError(t, err)

		repo := NewSQLiteRepository(gormDB)

		stocks := []Stock{
			{Ticker: "PETR4", Name: "Petrobras", Sector: "Energia", Rank: 10},
			{Ticker: "VALE3", Name: "Vale", Sector: "Mineração", Rank: 10},
		}

		for _, stock := range stocks {
			err = repo.Create(context.Background(), &stock)
			require.NoError(t, err)
		}

		list, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, list, 2)
		assert.Equal(t, "PETR4", list[0].Ticker)
		assert.Equal(t, "VALE3", list[1].Ticker)
	})

	t.Run("success - returns empty list when no stocks", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		err = gormDB.AutoMigrate(&Stock{})
		require.NoError(t, err)

		repo := NewSQLiteRepository(gormDB)

		list, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, list, 0)
	})
}

func TestSQLiteRepository_FindByTicker(t *testing.T) {
	t.Run("success - finds stock by ticker", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		err = gormDB.AutoMigrate(&Stock{})
		require.NoError(t, err)

		repo := NewSQLiteRepository(gormDB)

		stock := &Stock{
			Ticker: "PETR4",
			Name:   "Petrobras",
			Sector: "Energia",
			Rank:   10,
		}

		err = repo.Create(context.Background(), stock)
		require.NoError(t, err)

		found, err := repo.FindByTicker(context.Background(), "PETR4")
		assert.NoError(t, err)
		assert.Equal(t, stock.ID, found.ID)
		assert.Equal(t, "PETR4", found.Ticker)
		assert.Equal(t, "Petrobras", found.Name)
	})

	t.Run("error - stock not found", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		err = gormDB.AutoMigrate(&Stock{})
		require.NoError(t, err)

		repo := NewSQLiteRepository(gormDB)

		_, err = repo.FindByTicker(context.Background(), "XXXX")
		assert.Error(t, err)
		assert.Equal(t, ErrStockNotFound, err)
	})
}

func TestSQLiteRepository_Seed(t *testing.T) {
	t.Run("success - seeds stocks without force (idempotent)", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		err = gormDB.AutoMigrate(&Stock{})
		require.NoError(t, err)

		repo := NewSQLiteRepository(gormDB)

		stocks := []Stock{
			{Ticker: "PETR4", Name: "Petrobras", Sector: "Energia", Rank: 10},
			{Ticker: "VALE3", Name: "Vale", Sector: "Mineração", Rank: 10},
		}

		inserted, updated, skipped, err := repo.Seed(context.Background(), stocks, false)
		assert.NoError(t, err)
		assert.Equal(t, 2, inserted)
		assert.Equal(t, 0, updated)
		assert.Equal(t, 0, skipped)

		list, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, list, 2)

		inserted2, updated2, skipped2, err := repo.Seed(context.Background(), stocks, false)
		assert.NoError(t, err)
		assert.Equal(t, 0, inserted2)
		assert.Equal(t, 0, updated2)
		assert.Equal(t, 2, skipped2)

		list2, err := repo.List(context.Background())
		assert.NoError(t, err)
		assert.Len(t, list2, 2)
	})

	t.Run("success - seeds stocks with force (update)", func(t *testing.T) {
		gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
		require.NoError(t, err)

		err = gormDB.AutoMigrate(&Stock{})
		require.NoError(t, err)

		repo := NewSQLiteRepository(gormDB)

		stocks := []Stock{
			{Ticker: "PETR4", Name: "Petrobras", Sector: "Energia", Rank: 10},
		}

		inserted, updated, skipped, err := repo.Seed(context.Background(), stocks, false)
		assert.NoError(t, err)
		assert.Equal(t, 1, inserted)
		assert.Equal(t, 0, updated)
		assert.Equal(t, 0, skipped)

		updatedStocks := []Stock{
			{Ticker: "PETR4", Name: "Petrobras SA", Sector: "Energia", Rank: 9},
		}

		inserted2, updated2, skipped2, err := repo.Seed(context.Background(), updatedStocks, true)
		assert.NoError(t, err)
		assert.Equal(t, 0, inserted2)
		assert.Equal(t, 1, updated2)
		assert.Equal(t, 0, skipped2)

		found, err := repo.FindByTicker(context.Background(), "PETR4")
		assert.NoError(t, err)
		assert.Equal(t, "Petrobras SA", found.Name)
		assert.Equal(t, int8(9), found.Rank)
	})
}
