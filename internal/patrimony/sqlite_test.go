package patrimony

import (
	"context"
	"testing"

	"github.com/opinedajr/micro-investing/internal/infrastructure/database"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupSQLiteRepository(t *testing.T) *SQLiteRepository {
	gormDB, err := database.NewMemoryDatabase(t).Connect(context.Background())
	require.NoError(t, err)

	err = gormDB.AutoMigrate(&Patrimony{})
	require.NoError(t, err)

	return NewSQLiteRepository(gormDB)
}

func TestSQLiteRepository_Create(t *testing.T) {
	t.Run("success - creates and retrieves patrimony", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		patrimony := &Patrimony{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   150000,
		}

		err := repo.Create(context.Background(), patrimony)
		require.NoError(t, err)
		assert.NotEmpty(t, patrimony.ID)

		found, err := repo.FindByID(context.Background(), patrimony.ID)
		assert.NoError(t, err)
		assert.Equal(t, patrimony.ID, found.ID)
		assert.Equal(t, "wallet-id", found.WalletID)
		assert.Equal(t, 2026, found.Year)
		assert.Equal(t, 7, found.Month)
		assert.Equal(t, TypeStocks, found.Type)
		assert.Equal(t, int64(150000), found.Amount)
	})

	t.Run("error - violates unique constraint", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		first := &Patrimony{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   100000,
		}
		err := repo.Create(context.Background(), first)
		require.NoError(t, err)

		second := &Patrimony{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   200000,
		}
		err = repo.Create(context.Background(), second)
		assert.Error(t, err)
	})
}

func TestSQLiteRepository_Update(t *testing.T) {
	t.Run("success - updates patrimony amount", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		patrimony := &Patrimony{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   100000,
		}
		err := repo.Create(context.Background(), patrimony)
		require.NoError(t, err)

		patrimony.Amount = 200000
		err = repo.Update(context.Background(), patrimony)
		require.NoError(t, err)

		found, err := repo.FindByID(context.Background(), patrimony.ID)
		assert.NoError(t, err)
		assert.Equal(t, int64(200000), found.Amount)
	})
}

func TestSQLiteRepository_FindByID(t *testing.T) {
	t.Run("error - returns error for non-existent id", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		_, err := repo.FindByID(context.Background(), "non-existent-id")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPatrimonyNotFound)
	})
}

func TestSQLiteRepository_FindByFilter(t *testing.T) {
	t.Run("success - filters by wallet", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		require.NoError(t, repo.Create(context.Background(), &Patrimony{
			WalletID: "wallet-a",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   100000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Patrimony{
			WalletID: "wallet-a",
			Year:     2026,
			Month:    7,
			Type:     TypeFIIs,
			Amount:   50000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Patrimony{
			WalletID: "wallet-b",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   30000,
		}))

		results, err := repo.FindByFilter(context.Background(), PatrimonyFilter{WalletID: "wallet-a"})
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("success - filters by type", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		require.NoError(t, repo.Create(context.Background(), &Patrimony{
			WalletID: "wallet-a",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   100000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Patrimony{
			WalletID: "wallet-a",
			Year:     2026,
			Month:    7,
			Type:     TypeFIIs,
			Amount:   50000,
		}))

		results, err := repo.FindByFilter(context.Background(), PatrimonyFilter{
			WalletID: "wallet-a",
			Type:     TypeFIIs,
		})
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, TypeFIIs, results[0].Type)
	})

	t.Run("success - filters by year and month", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		require.NoError(t, repo.Create(context.Background(), &Patrimony{
			WalletID: "wallet-a",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   100000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Patrimony{
			WalletID: "wallet-a",
			Year:     2026,
			Month:    8,
			Type:     TypeStocks,
			Amount:   110000,
		}))

		results, err := repo.FindByFilter(context.Background(), PatrimonyFilter{
			WalletID: "wallet-a",
			Year:     2026,
			Month:    8,
		})
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, 8, results[0].Month)
	})
}

func TestSQLiteRepository_FindByWalletYearMonthType(t *testing.T) {
	t.Run("success - finds patrimony by unique combination", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		patrimony := &Patrimony{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   100000,
		}
		err := repo.Create(context.Background(), patrimony)
		require.NoError(t, err)

		found, err := repo.FindByWalletYearMonthType(context.Background(), "wallet-id", 2026, 7, TypeStocks)
		assert.NoError(t, err)
		assert.Equal(t, patrimony.ID, found.ID)
	})

	t.Run("error - returns record not found", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		_, err := repo.FindByWalletYearMonthType(context.Background(), "wallet-id", 2026, 7, TypeStocks)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrPatrimonyNotFound)
	})
}
