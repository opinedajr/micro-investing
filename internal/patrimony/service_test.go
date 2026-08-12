package patrimony

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestService_Create(t *testing.T) {
	t.Run("success - creates patrimony and returns output", func(t *testing.T) {
		now := time.Now()
		repo := &mockPatrimonyRepository{
			createFunc: func(ctx context.Context, patrimony *Patrimony) error {
				patrimony.ID = "patrimony-id"
				patrimony.CreatedAt = now
				patrimony.UpdatedAt = now
				return nil
			},
			findByWalletYearMonthTypeFn: func(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error) {
				return nil, ErrPatrimonyNotFound
			},
		}

		service := NewService(repo, &mockAssetRepository{})
		input := CreatePatrimonyInput{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   150000,
		}

		output, err := service.Create(context.Background(), input)

		assert.NoError(t, err)
		assert.Equal(t, "patrimony-id", output.ID)
		assert.Equal(t, "wallet-id", output.WalletID)
		assert.Equal(t, 2026, output.Year)
		assert.Equal(t, 7, output.Month)
		assert.Equal(t, TypeStocks, output.Type)
		assert.Equal(t, int64(150000), output.Amount)
	})

	t.Run("error - returns error when patrimony already exists", func(t *testing.T) {
		repo := &mockPatrimonyRepository{
			findByWalletYearMonthTypeFn: func(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error) {
				return &Patrimony{
					ID:       "existing-id",
					WalletID: walletID,
					Year:     year,
					Month:    month,
					Type:     assetType,
				}, nil
			},
		}

		service := NewService(repo, &mockAssetRepository{})
		input := CreatePatrimonyInput{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   150000,
		}

		_, err := service.Create(context.Background(), input)

		assert.Error(t, err)
		assert.Equal(t, ErrPatrimonyAlreadyExists, err)
	})

	t.Run("error - returns error for invalid asset type", func(t *testing.T) {
		repo := &mockPatrimonyRepository{}
		service := NewService(repo, &mockAssetRepository{})
		input := CreatePatrimonyInput{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     "invalid",
			Amount:   150000,
		}

		_, err := service.Create(context.Background(), input)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidAssetType, err)
	})

	t.Run("error - returns error for invalid year", func(t *testing.T) {
		repo := &mockPatrimonyRepository{}
		service := NewService(repo, &mockAssetRepository{})
		input := CreatePatrimonyInput{
			WalletID: "wallet-id",
			Year:     1999,
			Month:    7,
			Type:     TypeStocks,
			Amount:   150000,
		}

		_, err := service.Create(context.Background(), input)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidPatrimonyYear, err)
	})

	t.Run("error - returns error for invalid month", func(t *testing.T) {
		repo := &mockPatrimonyRepository{}
		service := NewService(repo, &mockAssetRepository{})
		input := CreatePatrimonyInput{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    13,
			Type:     TypeStocks,
			Amount:   150000,
		}

		_, err := service.Create(context.Background(), input)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidPatrimonyMonth, err)
	})

	t.Run("error - returns error for negative amount", func(t *testing.T) {
		repo := &mockPatrimonyRepository{}
		service := NewService(repo, &mockAssetRepository{})
		input := CreatePatrimonyInput{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   -1,
		}

		_, err := service.Create(context.Background(), input)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidPatrimonyAmount, err)
	})

	t.Run("error - returns error when repository fails", func(t *testing.T) {
		repo := &mockPatrimonyRepository{
			createFunc: func(ctx context.Context, patrimony *Patrimony) error {
				return errors.New("database error")
			},
			findByWalletYearMonthTypeFn: func(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error) {
				return nil, ErrPatrimonyNotFound
			},
		}

		service := NewService(repo, &mockAssetRepository{})
		input := CreatePatrimonyInput{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   150000,
		}

		_, err := service.Create(context.Background(), input)

		assert.Error(t, err)
	})
}

func TestService_Update(t *testing.T) {
	t.Run("success - updates patrimony", func(t *testing.T) {
		now := time.Now()
		existing := &Patrimony{
			ID:       "patrimony-id",
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   100000,
		}

		repo := &mockPatrimonyRepository{
			findByIDFunc: func(ctx context.Context, id string) (*Patrimony, error) {
				return existing, nil
			},
			findByWalletYearMonthTypeFn: func(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error) {
				return nil, ErrPatrimonyNotFound
			},
			updateFunc: func(ctx context.Context, patrimony *Patrimony) error {
				existing.Amount = patrimony.Amount
				existing.UpdatedAt = now
				return nil
			},
		}

		service := NewService(repo, &mockAssetRepository{})
		input := UpdatePatrimonyInput{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   200000,
		}

		output, err := service.Update(context.Background(), "patrimony-id", input)

		assert.NoError(t, err)
		assert.Equal(t, "patrimony-id", output.ID)
		assert.Equal(t, int64(200000), output.Amount)
	})

	t.Run("error - returns error when patrimony not found", func(t *testing.T) {
		repo := &mockPatrimonyRepository{
			findByIDFunc: func(ctx context.Context, id string) (*Patrimony, error) {
				return nil, ErrPatrimonyNotFound
			},
		}

		service := NewService(repo, &mockAssetRepository{})
		input := UpdatePatrimonyInput{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   200000,
		}

		_, err := service.Update(context.Background(), "non-existent", input)

		assert.Error(t, err)
		assert.Equal(t, ErrPatrimonyNotFound, err)
	})

	t.Run("error - returns error when wallet id does not match", func(t *testing.T) {
		repo := &mockPatrimonyRepository{
			findByIDFunc: func(ctx context.Context, id string) (*Patrimony, error) {
				return &Patrimony{
					ID:       "patrimony-id",
					WalletID: "other-wallet",
					Year:     2026,
					Month:    7,
					Type:     TypeStocks,
				}, nil
			},
		}

		service := NewService(repo, &mockAssetRepository{})
		input := UpdatePatrimonyInput{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   200000,
		}

		_, err := service.Update(context.Background(), "patrimony-id", input)

		assert.Error(t, err)
		assert.Equal(t, ErrPatrimonyNotFound, err)
	})

	t.Run("error - returns error when changing to existing combination", func(t *testing.T) {
		existing := &Patrimony{
			ID:       "patrimony-id",
			WalletID: "wallet-id",
			Year:     2026,
			Month:    7,
			Type:     TypeStocks,
			Amount:   100000,
		}

		repo := &mockPatrimonyRepository{
			findByIDFunc: func(ctx context.Context, id string) (*Patrimony, error) {
				return existing, nil
			},
			findByWalletYearMonthTypeFn: func(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error) {
				return &Patrimony{
					ID:       "other-id",
					WalletID: walletID,
					Year:     year,
					Month:    month,
					Type:     assetType,
				}, nil
			},
		}

		service := NewService(repo, &mockAssetRepository{})
		input := UpdatePatrimonyInput{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    8,
			Type:     TypeFIIs,
			Amount:   200000,
		}

		_, err := service.Update(context.Background(), "patrimony-id", input)

		assert.Error(t, err)
		assert.Equal(t, ErrPatrimonyAlreadyExists, err)
	})

	t.Run("error - returns error for invalid input", func(t *testing.T) {
		repo := &mockPatrimonyRepository{}
		service := NewService(repo, &mockAssetRepository{})
		input := UpdatePatrimonyInput{
			WalletID: "wallet-id",
			Year:     2026,
			Month:    0,
			Type:     TypeStocks,
			Amount:   200000,
		}

		_, err := service.Update(context.Background(), "patrimony-id", input)

		assert.Error(t, err)
	})
}

func TestService_List(t *testing.T) {
	t.Run("success - lists patrimonies for wallet", func(t *testing.T) {
		now := time.Now()
		repo := &mockPatrimonyRepository{
			findByFilterFunc: func(ctx context.Context, filter PatrimonyFilter) ([]Patrimony, error) {
				return []Patrimony{
					{
						ID:        "patrimony-1",
						WalletID:  filter.WalletID,
						Year:      2026,
						Month:     7,
						Type:      TypeStocks,
						Amount:    100000,
						CreatedAt: now,
						UpdatedAt: now,
					},
					{
						ID:        "patrimony-2",
						WalletID:  filter.WalletID,
						Year:      2026,
						Month:     7,
						Type:      TypeFIIs,
						Amount:    50000,
						CreatedAt: now,
						UpdatedAt: now,
					},
				}, nil
			},
		}

		service := NewService(repo, &mockAssetRepository{})
		filter := PatrimonyFilter{WalletID: "wallet-id"}

		outputs, err := service.List(context.Background(), filter)

		assert.NoError(t, err)
		assert.Len(t, outputs, 2)
		assert.Equal(t, "patrimony-1", outputs[0].ID)
		assert.Equal(t, TypeStocks, outputs[0].Type)
		assert.Equal(t, "patrimony-2", outputs[1].ID)
		assert.Equal(t, TypeFIIs, outputs[1].Type)
	})

	t.Run("error - returns error when repository fails", func(t *testing.T) {
		repo := &mockPatrimonyRepository{
			findByFilterFunc: func(ctx context.Context, filter PatrimonyFilter) ([]Patrimony, error) {
				return nil, errors.New("database error")
			},
		}

		service := NewService(repo, &mockAssetRepository{})
		filter := PatrimonyFilter{WalletID: "wallet-id"}

		_, err := service.List(context.Background(), filter)

		assert.Error(t, err)
	})
}
