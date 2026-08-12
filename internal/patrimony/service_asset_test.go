package patrimony

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestService_CreateAsset(t *testing.T) {
	t.Run("success - creates asset and recalculates patrimony in transaction", func(t *testing.T) {
		now := time.Now()
		date := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		}
		assetRepo := &mockAssetRepository{
			createFunc: func(ctx context.Context, asset *Asset) error {
				asset.ID = "asset-id"
				asset.CreatedAt = now
				asset.UpdatedAt = now
				return nil
			},
			sumFunc: func(ctx context.Context, walletID string, assetType AssetType, year int, month int) (int64, error) {
				return 150000, nil
			},
		}

		service := NewService(patrimonyRepo, assetRepo)
		input := CreateAssetInput{
			WalletID:    "wallet-id",
			Type:        TypeStocks,
			Date:        date.Format(time.RFC3339),
			Description: "PETR4 - Petrobras",
			Amount:      150000,
		}

		output, err := service.CreateAsset(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "asset-id", output.ID)
		assert.Equal(t, "wallet-id", output.WalletID)
		assert.Equal(t, TypeStocks, output.Type)
		assert.Equal(t, int64(150000), output.Amount)
		assert.Equal(t, "PETR4 - Petrobras", output.Description)
	})

	t.Run("success - creates patrimony when it does not exist (recalc creates new)", func(t *testing.T) {
		date := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
			findByWalletYearMonthTypeFn: func(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error) {
				return nil, ErrPatrimonyNotFound
			},
			createFunc: func(ctx context.Context, patrimony *Patrimony) error {
				patrimony.ID = "new-patrimony-id"
				return nil
			},
		}
		assetRepo := &mockAssetRepository{
			createFunc: func(ctx context.Context, asset *Asset) error {
				asset.ID = "asset-id"
				return nil
			},
			sumFunc: func(ctx context.Context, walletID string, assetType AssetType, year int, month int) (int64, error) {
				return 50000, nil
			},
		}

		service := NewService(patrimonyRepo, assetRepo)
		input := CreateAssetInput{
			WalletID:    "wallet-id",
			Type:        TypeStocks,
			Date:        date.Format(time.RFC3339),
			Description: "First asset",
			Amount:      50000,
		}

		output, err := service.CreateAsset(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
	})

	t.Run("success - updates existing patrimony when it exists", func(t *testing.T) {
		date := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
		updatedPatrimony := false

		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
			findByWalletYearMonthTypeFn: func(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error) {
				return &Patrimony{
					ID:       "existing-id",
					WalletID: walletID,
					Year:     year,
					Month:    month,
					Type:     assetType,
					Amount:   50000,
				}, nil
			},
			updateFunc: func(ctx context.Context, patrimony *Patrimony) error {
				updatedPatrimony = true
				return nil
			},
		}
		assetRepo := &mockAssetRepository{
			createFunc: func(ctx context.Context, asset *Asset) error {
				asset.ID = "asset-id"
				return nil
			},
			sumFunc: func(ctx context.Context, walletID string, assetType AssetType, year int, month int) (int64, error) {
				return 75000, nil
			},
		}

		service := NewService(patrimonyRepo, assetRepo)
		input := CreateAssetInput{
			WalletID:    "wallet-id",
			Type:        TypeStocks,
			Date:        date.Format(time.RFC3339),
			Description: "Second asset",
			Amount:      25000,
		}

		output, err := service.CreateAsset(context.Background(), input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.True(t, updatedPatrimony)
	})

	t.Run("error - returns error for invalid date format", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		}
		assetRepo := &mockAssetRepository{}

		service := NewService(patrimonyRepo, assetRepo)
		input := CreateAssetInput{
			WalletID:    "wallet-id",
			Type:        TypeStocks,
			Date:        "invalid-date",
			Description: "Description valid",
			Amount:      150000,
		}

		_, err := service.CreateAsset(context.Background(), input)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAssetDate, err)
	})

	t.Run("error - returns error for future date", func(t *testing.T) {
		futureDate := time.Now().Add(24 * time.Hour)

		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		}
		assetRepo := &mockAssetRepository{}

		service := NewService(patrimonyRepo, assetRepo)
		input := CreateAssetInput{
			WalletID:    "wallet-id",
			Type:        TypeStocks,
			Date:        futureDate.Format(time.RFC3339),
			Description: "Description valid",
			Amount:      150000,
		}

		_, err := service.CreateAsset(context.Background(), input)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAssetDate, err)
	})

	t.Run("error - returns error for invalid asset type", func(t *testing.T) {
		date := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		}
		assetRepo := &mockAssetRepository{}

		service := NewService(patrimonyRepo, assetRepo)
		input := CreateAssetInput{
			WalletID:    "wallet-id",
			Type:        "invalid",
			Date:        date.Format(time.RFC3339),
			Description: "Description valid",
			Amount:      150000,
		}

		_, err := service.CreateAsset(context.Background(), input)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAssetType, err)
	})

	t.Run("error - returns error for description too short", func(t *testing.T) {
		date := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		}
		assetRepo := &mockAssetRepository{}

		service := NewService(patrimonyRepo, assetRepo)
		input := CreateAssetInput{
			WalletID:    "wallet-id",
			Type:        TypeStocks,
			Date:        date.Format(time.RFC3339),
			Description: "ab",
			Amount:      150000,
		}

		_, err := service.CreateAsset(context.Background(), input)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAssetDescription, err)
	})

	t.Run("error - returns error for amount <= 0", func(t *testing.T) {
		date := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		}
		assetRepo := &mockAssetRepository{}

		service := NewService(patrimonyRepo, assetRepo)
		input := CreateAssetInput{
			WalletID:    "wallet-id",
			Type:        TypeStocks,
			Date:        date.Format(time.RFC3339),
			Description: "Valid description",
			Amount:      0,
		}

		_, err := service.CreateAsset(context.Background(), input)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAssetAmount, err)
	})

	t.Run("error - returns error when asset repository fails", func(t *testing.T) {
		date := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		}
		assetRepo := &mockAssetRepository{
			createFunc: func(ctx context.Context, asset *Asset) error {
				return errors.New("database error")
			},
		}

		service := NewService(patrimonyRepo, assetRepo)
		input := CreateAssetInput{
			WalletID:    "wallet-id",
			Type:        TypeStocks,
			Date:        date.Format(time.RFC3339),
			Description: "Valid description",
			Amount:      150000,
		}

		_, err := service.CreateAsset(context.Background(), input)

		assert.Error(t, err)
	})
}

func TestService_DeleteAsset(t *testing.T) {
	t.Run("success - deletes asset and recalculates patrimony", func(t *testing.T) {
		deleted := false
		date := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
			findByWalletYearMonthTypeFn: func(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error) {
				return &Patrimony{
					ID:       "patrimony-id",
					WalletID: walletID,
					Year:     year,
					Month:    month,
					Type:     assetType,
					Amount:   150000,
				}, nil
			},
			updateFunc: func(ctx context.Context, patrimony *Patrimony) error {
				return nil
			},
		}
		assetRepo := &mockAssetRepository{
			findByIDFunc: func(ctx context.Context, id string) (*Asset, error) {
				return &Asset{
					ID:       id,
					WalletID: "wallet-id",
					Type:     TypeStocks,
					Date:     date,
					Amount:   50000,
				}, nil
			},
			deleteFunc: func(ctx context.Context, id string) error {
				deleted = true
				return nil
			},
			sumFunc: func(ctx context.Context, walletID string, assetType AssetType, year int, month int) (int64, error) {
				return 100000, nil
			},
		}

		service := NewService(patrimonyRepo, assetRepo)

		err := service.DeleteAsset(context.Background(), "asset-id")

		assert.NoError(t, err)
		assert.True(t, deleted)
	})

	t.Run("success - deletes patrimony if sum becomes zero", func(t *testing.T) {
		date := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
			findByWalletYearMonthTypeFn: func(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error) {
				return &Patrimony{
					ID:       "patrimony-id",
					WalletID: walletID,
					Year:     year,
					Month:    month,
					Type:     assetType,
					Amount:   50000,
				}, nil
			},
			updateFunc: func(ctx context.Context, patrimony *Patrimony) error {
				return nil
			},
		}
		assetRepo := &mockAssetRepository{
			findByIDFunc: func(ctx context.Context, id string) (*Asset, error) {
				return &Asset{
					ID:       id,
					WalletID: "wallet-id",
					Type:     TypeStocks,
					Date:     date,
					Amount:   50000,
				}, nil
			},
			deleteFunc: func(ctx context.Context, id string) error {
				return nil
			},
			sumFunc: func(ctx context.Context, walletID string, assetType AssetType, year int, month int) (int64, error) {
				return 0, nil
			},
		}

		service := NewService(patrimonyRepo, assetRepo)

		err := service.DeleteAsset(context.Background(), "asset-id")

		assert.NoError(t, err)
	})

	t.Run("error - returns error when asset not found", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		}
		assetRepo := &mockAssetRepository{
			findByIDFunc: func(ctx context.Context, id string) (*Asset, error) {
				return nil, ErrAssetNotFound
			},
		}

		service := NewService(patrimonyRepo, assetRepo)

		err := service.DeleteAsset(context.Background(), "missing-id")

		assert.Error(t, err)
		assert.Equal(t, ErrAssetNotFound, err)
	})

	t.Run("error - returns error when delete fails", func(t *testing.T) {
		date := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)

		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		}
		assetRepo := &mockAssetRepository{
			findByIDFunc: func(ctx context.Context, id string) (*Asset, error) {
				return &Asset{
					ID:       id,
					WalletID: "wallet-id",
					Type:     TypeStocks,
					Date:     date,
					Amount:   50000,
				}, nil
			},
			deleteFunc: func(ctx context.Context, id string) error {
				return errors.New("database error")
			},
		}

		service := NewService(patrimonyRepo, assetRepo)

		err := service.DeleteAsset(context.Background(), "asset-id")

		assert.Error(t, err)
	})
}
