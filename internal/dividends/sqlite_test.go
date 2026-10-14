package dividends

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

	require.NoError(t, gormDB.AutoMigrate(&Dividend{}))

	return NewSQLiteRepository(gormDB)
}

func TestSQLiteRepository_Create(t *testing.T) {
	t.Run("success - creates and retrieves dividend", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		dividend := &Dividend{
			WalletID: "wallet-id",
			Year:     2026,
			Amount:   150000,
		}

		err := repo.Create(context.Background(), dividend)
		require.NoError(t, err)
		assert.NotEmpty(t, dividend.ID)

		found, err := repo.FindByWalletYear(context.Background(), "wallet-id", 2026)
		require.NoError(t, err)
		assert.Equal(t, dividend.ID, found.ID)
		assert.Equal(t, "wallet-id", found.WalletID)
		assert.Equal(t, 2026, found.Year)
		assert.Equal(t, int64(150000), found.Amount)
	})

	t.Run("error - violates unique constraint", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		first := &Dividend{WalletID: "wallet-id", Year: 2026, Amount: 100000}
		err := repo.Create(context.Background(), first)
		require.NoError(t, err)

		second := &Dividend{WalletID: "wallet-id", Year: 2026, Amount: 200000}
		err = repo.Create(context.Background(), second)
		assert.Error(t, err)
	})
}

func TestSQLiteRepository_FindByWalletYear(t *testing.T) {
	t.Run("success - finds dividend by wallet and year", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		dividend := &Dividend{WalletID: "wallet-id", Year: 2025, Amount: 90000}
		require.NoError(t, repo.Create(context.Background(), dividend))

		found, err := repo.FindByWalletYear(context.Background(), "wallet-id", 2025)
		require.NoError(t, err)
		assert.Equal(t, dividend.ID, found.ID)
	})

	t.Run("error - not found", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		found, err := repo.FindByWalletYear(context.Background(), "wallet-id", 2030)
		assert.Nil(t, found)
		assert.ErrorIs(t, err, ErrDividendNotFound)
	})

	t.Run("success - same year on another wallet does not conflict", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		require.NoError(t, repo.Create(context.Background(), &Dividend{WalletID: "wallet-a", Year: 2026, Amount: 1000}))
		require.NoError(t, repo.Create(context.Background(), &Dividend{WalletID: "wallet-b", Year: 2026, Amount: 2000}))

		found, err := repo.FindByWalletYear(context.Background(), "wallet-b", 2026)
		require.NoError(t, err)
		assert.Equal(t, int64(2000), found.Amount)
	})
}

func TestSQLiteRepository_FindByFilter(t *testing.T) {
	t.Run("success - returns dividends ordered by year desc", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		require.NoError(t, repo.Create(context.Background(), &Dividend{WalletID: "wallet-id", Year: 2024, Amount: 50000}))
		require.NoError(t, repo.Create(context.Background(), &Dividend{WalletID: "wallet-id", Year: 2026, Amount: 150000}))
		require.NoError(t, repo.Create(context.Background(), &Dividend{WalletID: "wallet-id", Year: 2025, Amount: 100000}))
		require.NoError(t, repo.Create(context.Background(), &Dividend{WalletID: "other-wallet", Year: 2026, Amount: 9999}))

		dividends, err := repo.FindByFilter(context.Background(), DividendFilter{WalletID: "wallet-id"})
		require.NoError(t, err)
		require.Len(t, dividends, 3)
		assert.Equal(t, 2026, dividends[0].Year)
		assert.Equal(t, 2025, dividends[1].Year)
		assert.Equal(t, 2024, dividends[2].Year)
	})

	t.Run("success - filters by year", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		require.NoError(t, repo.Create(context.Background(), &Dividend{WalletID: "wallet-id", Year: 2025, Amount: 100000}))
		require.NoError(t, repo.Create(context.Background(), &Dividend{WalletID: "wallet-id", Year: 2026, Amount: 150000}))

		dividends, err := repo.FindByFilter(context.Background(), DividendFilter{WalletID: "wallet-id", Year: 2025})
		require.NoError(t, err)
		require.Len(t, dividends, 1)
		assert.Equal(t, 2025, dividends[0].Year)
	})

	t.Run("success - filter without match returns empty list", func(t *testing.T) {
		repo := setupSQLiteRepository(t)

		require.NoError(t, repo.Create(context.Background(), &Dividend{WalletID: "wallet-id", Year: 2026, Amount: 150000}))

		dividends, err := repo.FindByFilter(context.Background(), DividendFilter{WalletID: "wallet-id", Year: 1990})
		require.NoError(t, err)
		assert.Empty(t, dividends)
	})
}
