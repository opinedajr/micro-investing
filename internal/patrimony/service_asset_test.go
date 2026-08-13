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

	t.Run("success - sets existing patrimony amount to zero when sum becomes zero", func(t *testing.T) {
		date := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
		updatedPatrimony := false
		var updatedAmount int64 = -1

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
				updatedPatrimony = true
				updatedAmount = patrimony.Amount
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
		assert.True(t, updatedPatrimony, "updateFunc must be invoked to persist the zeroed patrimony amount")
		assert.Equal(t, int64(0), updatedAmount, "existing patrimony amount must be set to zero when sum is zero (spec: patrimony is never deleted)")
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

func TestService_UpdateAsset(t *testing.T) {
	t.Run("success - updates asset and recalculates patrimony when month/type unchanged", func(t *testing.T) {
		now := time.Now()
		oldDate := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
		newDate := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)

		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		}
		assetRepo := &mockAssetRepository{
			findByIDFunc: func(ctx context.Context, id string) (*Asset, error) {
				return &Asset{
					ID:          id,
					WalletID:    "wallet-id",
					Type:        TypeStocks,
					Date:        oldDate,
					Description: "Old description",
					Amount:      50000,
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil
			},
			updateFunc: func(ctx context.Context, asset *Asset) error {
				return nil
			},
			sumFunc: func(ctx context.Context, walletID string, assetType AssetType, year int, month int) (int64, error) {
				return 75000, nil
			},
		}

		service := NewService(patrimonyRepo, assetRepo)
		input := UpdateAssetInput{
			WalletID:    "wallet-id",
			Type:        TypeStocks,
			Date:        newDate.Format(time.RFC3339),
			Description: "Updated description",
			Amount:      75000,
		}

		output, err := service.UpdateAsset(context.Background(), "asset-id", input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, "asset-id", output.ID)
		assert.Equal(t, int64(75000), output.Amount)
		assert.Equal(t, "Updated description", output.Description)
	})

	t.Run("success - recalculates both old and new month when asset date changes month", func(t *testing.T) {
		oldDate := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
		newDate := time.Date(2026, 8, 10, 12, 0, 0, 0, time.UTC)

		patrimonyRecalcCalls := []struct {
			walletID string
			typ      AssetType
			year     int
			month    int
		}{}

		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
			findByWalletYearMonthTypeFn: func(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error) {
				return nil, ErrPatrimonyNotFound
			},
		}
		assetRepo := &mockAssetRepository{
			findByIDFunc: func(ctx context.Context, id string) (*Asset, error) {
				return &Asset{
					ID:       id,
					WalletID: "wallet-id",
					Type:     TypeStocks,
					Date:     oldDate,
				}, nil
			},
			updateFunc: func(ctx context.Context, asset *Asset) error {
				return nil
			},
			sumFunc: func(ctx context.Context, walletID string, assetType AssetType, year int, month int) (int64, error) {
				patrimonyRecalcCalls = append(patrimonyRecalcCalls, struct {
					walletID string
					typ      AssetType
					year     int
					month    int
				}{walletID, assetType, year, month})
				return 0, nil
			},
		}

		service := NewService(patrimonyRepo, assetRepo)
		input := UpdateAssetInput{
			WalletID:    "wallet-id",
			Type:        TypeStocks,
			Date:        newDate.Format(time.RFC3339),
			Description: "Updated description",
			Amount:      100000,
		}

		output, err := service.UpdateAsset(context.Background(), "asset-id", input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Len(t, patrimonyRecalcCalls, 2, "should recalculate both old and new month patrimony")
		assert.Equal(t, 2026, patrimonyRecalcCalls[0].year)
		assert.Equal(t, 7, patrimonyRecalcCalls[0].month)
		assert.Equal(t, 2026, patrimonyRecalcCalls[1].year)
		assert.Equal(t, 8, patrimonyRecalcCalls[1].month)
	})

	t.Run("success - recalculates both old and new patrimony when asset type changes", func(t *testing.T) {
		oldDate := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
		newDate := time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC)

		patrimonyRecalcCalls := []AssetType{}

		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
			findByWalletYearMonthTypeFn: func(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error) {
				return nil, ErrPatrimonyNotFound
			},
		}
		assetRepo := &mockAssetRepository{
			findByIDFunc: func(ctx context.Context, id string) (*Asset, error) {
				return &Asset{
					ID:       id,
					WalletID: "wallet-id",
					Type:     TypeStocks,
					Date:     oldDate,
				}, nil
			},
			updateFunc: func(ctx context.Context, asset *Asset) error {
				return nil
			},
			sumFunc: func(ctx context.Context, walletID string, assetType AssetType, year int, month int) (int64, error) {
				patrimonyRecalcCalls = append(patrimonyRecalcCalls, assetType)
				return 0, nil
			},
		}

		service := NewService(patrimonyRepo, assetRepo)
		input := UpdateAssetInput{
			WalletID:    "wallet-id",
			Type:        TypeFIIs,
			Date:        newDate.Format(time.RFC3339),
			Description: "Updated description",
			Amount:      100000,
		}

		output, err := service.UpdateAsset(context.Background(), "asset-id", input)

		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Len(t, patrimonyRecalcCalls, 2)
		assert.Equal(t, TypeStocks, patrimonyRecalcCalls[0])
		assert.Equal(t, TypeFIIs, patrimonyRecalcCalls[1])
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
		date := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
		input := UpdateAssetInput{
			WalletID:    "wallet-id",
			Type:        TypeStocks,
			Date:        date.Format(time.RFC3339),
			Description: "Valid description",
			Amount:      150000,
		}

		_, err := service.UpdateAsset(context.Background(), "missing-id", input)

		assert.Error(t, err)
		assert.Equal(t, ErrAssetNotFound, err)
	})

	t.Run("error - returns error for invalid date format", func(t *testing.T) {
		patrimonyRepo := &mockPatrimonyRepository{
			runInTransactionFn: func(ctx context.Context, fn func(ctx context.Context) error) error {
				return fn(ctx)
			},
		}
		assetRepo := &mockAssetRepository{}

		service := NewService(patrimonyRepo, assetRepo)
		input := UpdateAssetInput{
			WalletID:    "wallet-id",
			Type:        TypeStocks,
			Date:        "invalid-date",
			Description: "Valid description",
			Amount:      150000,
		}

		_, err := service.UpdateAsset(context.Background(), "asset-id", input)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAssetDate, err)
	})
}

func TestService_ListAssets(t *testing.T) {
	t.Run("success - lists assets for wallet", func(t *testing.T) {
		assetRepo := &mockAssetRepository{
			findByFilterFunc: func(ctx context.Context, filter AssetFilter) ([]Asset, error) {
				return []Asset{
					{
						ID:          "asset-1",
						WalletID:    filter.WalletID,
						Type:        TypeStocks,
						Date:        time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC),
						Description: "Asset 1",
						Amount:      100000,
					},
					{
						ID:          "asset-2",
						WalletID:    filter.WalletID,
						Type:        TypeFIIs,
						Date:        time.Date(2026, 7, 20, 12, 0, 0, 0, time.UTC),
						Description: "Asset 2",
						Amount:      50000,
					},
				}, nil
			},
		}

		service := NewService(&mockPatrimonyRepository{}, assetRepo)
		filter := AssetFilter{WalletID: "wallet-id"}

		outputs, err := service.ListAssets(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, outputs, 2)
		assert.Equal(t, "asset-1", outputs[0].ID)
		assert.Equal(t, TypeStocks, outputs[0].Type)
	})

	t.Run("success - returns empty list when no assets found", func(t *testing.T) {
		assetRepo := &mockAssetRepository{
			findByFilterFunc: func(ctx context.Context, filter AssetFilter) ([]Asset, error) {
				return []Asset{}, nil
			},
		}

		service := NewService(&mockPatrimonyRepository{}, assetRepo)
		filter := AssetFilter{WalletID: "wallet-id"}

		outputs, err := service.ListAssets(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, outputs, 0)
	})

	t.Run("error - returns error when start_date is after end_date", func(t *testing.T) {
		start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)

		assetRepo := &mockAssetRepository{}

		service := NewService(&mockPatrimonyRepository{}, assetRepo)
		filter := AssetFilter{
			WalletID:  "wallet-id",
			StartDate: &start,
			EndDate:   &end,
		}

		_, err := service.ListAssets(context.Background(), filter)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidDateRange, err)
	})

	t.Run("success - start_date equal to end_date is allowed", func(t *testing.T) {
		sameDate := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)

		assetRepo := &mockAssetRepository{
			findByFilterFunc: func(ctx context.Context, filter AssetFilter) ([]Asset, error) {
				return []Asset{
					{
						ID:          "asset-1",
						WalletID:    filter.WalletID,
						Type:        TypeStocks,
						Date:        time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC),
						Description: "Asset 1",
						Amount:      100000,
					},
				}, nil
			},
		}

		service := NewService(&mockPatrimonyRepository{}, assetRepo)
		filter := AssetFilter{
			WalletID:  "wallet-id",
			StartDate: &sameDate,
			EndDate:   &sameDate,
		}

		outputs, err := service.ListAssets(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, outputs, 1)
	})

	t.Run("error - returns error when repository fails", func(t *testing.T) {
		assetRepo := &mockAssetRepository{
			findByFilterFunc: func(ctx context.Context, filter AssetFilter) ([]Asset, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(&mockPatrimonyRepository{}, assetRepo)
		filter := AssetFilter{WalletID: "wallet-id"}

		_, err := service.ListAssets(context.Background(), filter)

		assert.Error(t, err)
	})
}
