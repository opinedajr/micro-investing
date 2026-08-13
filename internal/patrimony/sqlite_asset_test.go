package patrimony

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSQLiteAssetRepository_Create(t *testing.T) {
	t.Run("success - creates and retrieves asset", func(t *testing.T) {
		repo := setupSQLiteAssetRepository(t)

		date := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
		asset := &Asset{
			WalletID:    "wallet-id",
			Type:        TypeStocks,
			Date:        date,
			Description: "PETR4 - Petrobras",
			Amount:      150000,
		}

		err := repo.Create(context.Background(), asset)
		require.NoError(t, err)
		assert.NotEmpty(t, asset.ID)

		found, err := repo.FindByID(context.Background(), asset.ID)
		assert.NoError(t, err)
		assert.Equal(t, "wallet-id", found.WalletID)
		assert.Equal(t, TypeStocks, found.Type)
		assert.Equal(t, date.Format(time.RFC3339), found.Date.Format(time.RFC3339))
		assert.Equal(t, "PETR4 - Petrobras", found.Description)
		assert.Equal(t, int64(150000), found.Amount)
	})
}

func TestSQLiteAssetRepository_Delete(t *testing.T) {
	t.Run("success - deletes asset", func(t *testing.T) {
		repo := setupSQLiteAssetRepository(t)

		asset := &Asset{
			WalletID:    "wallet-id",
			Type:        TypeStocks,
			Date:        time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC),
			Description: "Asset description",
			Amount:      150000,
		}
		require.NoError(t, repo.Create(context.Background(), asset))

		err := repo.Delete(context.Background(), asset.ID)
		assert.NoError(t, err)

		_, err = repo.FindByID(context.Background(), asset.ID)
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrAssetNotFound)
	})

	t.Run("error - returns error for non-existent id", func(t *testing.T) {
		repo := setupSQLiteAssetRepository(t)

		err := repo.Delete(context.Background(), "non-existent-id")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrAssetNotFound)
	})
}

func TestSQLiteAssetRepository_FindByID(t *testing.T) {
	t.Run("error - returns error for non-existent id", func(t *testing.T) {
		repo := setupSQLiteAssetRepository(t)

		_, err := repo.FindByID(context.Background(), "non-existent-id")
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrAssetNotFound)
	})
}

func TestSQLiteAssetRepository_FindByFilter(t *testing.T) {
	t.Run("success - filters by wallet", func(t *testing.T) {
		repo := setupSQLiteAssetRepository(t)

		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC),
			Description: "Asset a1",
			Amount:      100000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeFIIs,
			Date:        time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC),
			Description: "Asset a2",
			Amount:      50000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-b",
			Type:        TypeStocks,
			Date:        time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC),
			Description: "Asset b1",
			Amount:      30000,
		}))

		results, err := repo.FindByFilter(context.Background(), AssetFilter{WalletID: "wallet-a"})
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("success - filters by type", func(t *testing.T) {
		repo := setupSQLiteAssetRepository(t)

		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC),
			Description: "Asset 1",
			Amount:      100000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeFIIs,
			Date:        time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC),
			Description: "Asset 2",
			Amount:      50000,
		}))

		results, err := repo.FindByFilter(context.Background(), AssetFilter{
			WalletID: "wallet-a",
			Type:     TypeFIIs,
		})
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, TypeFIIs, results[0].Type)
	})

	t.Run("success - filters by start_date inclusive", func(t *testing.T) {
		repo := setupSQLiteAssetRepository(t)

		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
			Description: "Asset 1",
			Amount:      100000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC),
			Description: "Asset 2",
			Amount:      50000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC),
			Description: "Asset 3",
			Amount:      75000,
		}))

		start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		results, err := repo.FindByFilter(context.Background(), AssetFilter{
			WalletID:  "wallet-a",
			StartDate: &start,
		})
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("success - filters by end_date inclusive", func(t *testing.T) {
		repo := setupSQLiteAssetRepository(t)

		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
			Description: "Asset 1",
			Amount:      100000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC),
			Description: "Asset 2",
			Amount:      50000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC),
			Description: "Asset 3",
			Amount:      75000,
		}))

		end := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
		results, err := repo.FindByFilter(context.Background(), AssetFilter{
			WalletID: "wallet-a",
			EndDate:  &end,
		})
		assert.NoError(t, err)
		assert.Len(t, results, 2)
	})

	t.Run("success - filters by start_date and end_date combined", func(t *testing.T) {
		repo := setupSQLiteAssetRepository(t)

		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC),
			Description: "Asset 1",
			Amount:      100000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC),
			Description: "Asset 2",
			Amount:      50000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 8, 15, 12, 0, 0, 0, time.UTC),
			Description: "Asset 3",
			Amount:      75000,
		}))

		start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
		results, err := repo.FindByFilter(context.Background(), AssetFilter{
			WalletID:  "wallet-a",
			StartDate: &start,
			EndDate:   &end,
		})
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, int64(50000), results[0].Amount)
	})

	t.Run("success - combines type and date range filters", func(t *testing.T) {
		repo := setupSQLiteAssetRepository(t)

		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC),
			Description: "Asset 1",
			Amount:      100000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeFIIs,
			Date:        time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC),
			Description: "Asset 2",
			Amount:      50000,
		}))

		start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 7, 31, 0, 0, 0, 0, time.UTC)
		results, err := repo.FindByFilter(context.Background(), AssetFilter{
			WalletID:  "wallet-a",
			Type:      TypeFIIs,
			StartDate: &start,
			EndDate:   &end,
		})
		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, TypeFIIs, results[0].Type)
	})
}

func TestSQLiteAssetRepository_SumByWalletTypeAndMonth(t *testing.T) {
	t.Run("success - sums assets by wallet, type, year and month", func(t *testing.T) {
		repo := setupSQLiteAssetRepository(t)

		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC),
			Description: "Asset 1",
			Amount:      100000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC),
			Description: "Asset 2",
			Amount:      50000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeStocks,
			Date:        time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC),
			Description: "Asset 3",
			Amount:      75000,
		}))
		require.NoError(t, repo.Create(context.Background(), &Asset{
			WalletID:    "wallet-a",
			Type:        TypeFIIs,
			Date:        time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC),
			Description: "Asset 4",
			Amount:      20000,
		}))

		total, err := repo.SumByWalletTypeAndMonth(context.Background(), "wallet-a", TypeStocks, 2026, 7)
		assert.NoError(t, err)
		assert.Equal(t, int64(150000), total)
	})

	t.Run("success - returns zero when no assets match", func(t *testing.T) {
		repo := setupSQLiteAssetRepository(t)

		total, err := repo.SumByWalletTypeAndMonth(context.Background(), "wallet-a", TypeStocks, 2026, 7)
		assert.NoError(t, err)
		assert.Equal(t, int64(0), total)
	})
}
