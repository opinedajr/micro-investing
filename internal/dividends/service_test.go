package dividends

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	currentYear := time.Now().Year()

	t.Run("success - creates dividend", func(t *testing.T) {
		repo := &mockRepository{
			createFunc: func(ctx context.Context, dividend *Dividend) error {
				dividend.ID = "dividend-id"
				return nil
			},
		}
		svc := NewService(repo)

		output, err := svc.Create(context.Background(), CreateDividendInput{
			WalletID: "wallet-id",
			Year:     currentYear,
			Amount:   150000,
		})

		require.NoError(t, err)
		assert.Equal(t, "dividend-id", output.ID)
		assert.Equal(t, currentYear, output.Year)
		assert.Equal(t, int64(150000), output.Amount)
	})

	t.Run("error - year below 1900", func(t *testing.T) {
		svc := NewService(&mockRepository{})

		_, err := svc.Create(context.Background(), CreateDividendInput{
			WalletID: "wallet-id",
			Year:     1899,
			Amount:   150000,
		})

		assert.ErrorIs(t, err, ErrInvalidDividendYear)
	})

	t.Run("error - year above current year plus one", func(t *testing.T) {
		svc := NewService(&mockRepository{})

		_, err := svc.Create(context.Background(), CreateDividendInput{
			WalletID: "wallet-id",
			Year:     time.Now().Year() + 2,
			Amount:   150000,
		})

		assert.ErrorIs(t, err, ErrInvalidDividendYear)
	})

	t.Run("error - amount not greater than zero", func(t *testing.T) {
		svc := NewService(&mockRepository{})

		_, err := svc.Create(context.Background(), CreateDividendInput{
			WalletID: "wallet-id",
			Year:     currentYear,
			Amount:   0,
		})

		assert.ErrorIs(t, err, ErrInvalidDividendAmount)
	})

	t.Run("error - duplicate year for wallet", func(t *testing.T) {
		repo := &mockRepository{
			findByWalletYearFunc: func(ctx context.Context, walletID string, year int) (*Dividend, error) {
				return &Dividend{ID: "existing-id", WalletID: walletID, Year: year}, nil
			},
		}
		svc := NewService(repo)

		_, err := svc.Create(context.Background(), CreateDividendInput{
			WalletID: "wallet-id",
			Year:     currentYear,
			Amount:   150000,
		})

		assert.ErrorIs(t, err, ErrDividendAlreadyExists)
	})

	t.Run("error - repository failure on lookup", func(t *testing.T) {
		repoErr := errors.New("db error")
		repo := &mockRepository{
			findByWalletYearFunc: func(ctx context.Context, walletID string, year int) (*Dividend, error) {
				return nil, repoErr
			},
		}
		svc := NewService(repo)

		_, err := svc.Create(context.Background(), CreateDividendInput{
			WalletID: "wallet-id",
			Year:     currentYear,
			Amount:   150000,
		})

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("error - repository failure on create", func(t *testing.T) {
		repoErr := errors.New("db error")
		repo := &mockRepository{
			createFunc: func(ctx context.Context, dividend *Dividend) error {
				return repoErr
			},
		}
		svc := NewService(repo)

		_, err := svc.Create(context.Background(), CreateDividendInput{
			WalletID: "wallet-id",
			Year:     currentYear,
			Amount:   150000,
		})

		assert.ErrorIs(t, err, repoErr)
	})
}

func TestService_Update(t *testing.T) {
	currentYear := time.Now().Year()

	t.Run("success - updates year and amount", func(t *testing.T) {
		repo := &mockRepository{
			findByIDFunc: func(ctx context.Context, walletID string, id string) (*Dividend, error) {
				return &Dividend{ID: id, WalletID: "wallet-id", Year: 2024, Amount: 100000}, nil
			},
			updateFunc: func(ctx context.Context, dividend *Dividend) error {
				return nil
			},
		}
		svc := NewService(repo)

		output, err := svc.Update(context.Background(), UpdateDividendInput{
			ID:       "dividend-id",
			WalletID: "wallet-id",
			Year:     currentYear,
			Amount:   250000,
		})

		require.NoError(t, err)
		assert.Equal(t, "dividend-id", output.ID)
		assert.Equal(t, currentYear, output.Year)
		assert.Equal(t, int64(250000), output.Amount)
	})

	t.Run("success - keeps same year without conflict", func(t *testing.T) {
		repo := &mockRepository{
			findByIDFunc: func(ctx context.Context, walletID string, id string) (*Dividend, error) {
				return &Dividend{ID: id, WalletID: "wallet-id", Year: 2024, Amount: 100000}, nil
			},
			findByWalletYearFunc: func(ctx context.Context, walletID string, year int) (*Dividend, error) {
				return &Dividend{ID: "dividend-id", WalletID: walletID, Year: year}, nil
			},
			updateFunc: func(ctx context.Context, dividend *Dividend) error {
				return nil
			},
		}
		svc := NewService(repo)

		output, err := svc.Update(context.Background(), UpdateDividendInput{
			ID:       "dividend-id",
			WalletID: "wallet-id",
			Year:     2024,
			Amount:   175000,
		})

		require.NoError(t, err)
		assert.Equal(t, int64(175000), output.Amount)
	})

	t.Run("error - dividend not found", func(t *testing.T) {
		repo := &mockRepository{}
		svc := NewService(repo)

		_, err := svc.Update(context.Background(), UpdateDividendInput{
			ID:       "missing-id",
			WalletID: "wallet-id",
			Year:     currentYear,
			Amount:   150000,
		})

		assert.ErrorIs(t, err, ErrDividendNotFound)
	})

	t.Run("error - year below 1900", func(t *testing.T) {
		svc := NewService(&mockRepository{})

		_, err := svc.Update(context.Background(), UpdateDividendInput{
			ID:       "dividend-id",
			WalletID: "wallet-id",
			Year:     1899,
			Amount:   150000,
		})

		assert.ErrorIs(t, err, ErrInvalidDividendYear)
	})

	t.Run("error - year above current year plus one", func(t *testing.T) {
		svc := NewService(&mockRepository{})

		_, err := svc.Update(context.Background(), UpdateDividendInput{
			ID:       "dividend-id",
			WalletID: "wallet-id",
			Year:     time.Now().Year() + 2,
			Amount:   150000,
		})

		assert.ErrorIs(t, err, ErrInvalidDividendYear)
	})

	t.Run("error - amount not greater than zero", func(t *testing.T) {
		svc := NewService(&mockRepository{})

		_, err := svc.Update(context.Background(), UpdateDividendInput{
			ID:       "dividend-id",
			WalletID: "wallet-id",
			Year:     currentYear,
			Amount:   0,
		})

		assert.ErrorIs(t, err, ErrInvalidDividendAmount)
	})

	t.Run("error - target year already taken by another dividend", func(t *testing.T) {
		repo := &mockRepository{
			findByIDFunc: func(ctx context.Context, walletID string, id string) (*Dividend, error) {
				return &Dividend{ID: id, WalletID: "wallet-id", Year: 2024, Amount: 100000}, nil
			},
			findByWalletYearFunc: func(ctx context.Context, walletID string, year int) (*Dividend, error) {
				return &Dividend{ID: "other-id", WalletID: walletID, Year: year}, nil
			},
		}
		svc := NewService(repo)

		_, err := svc.Update(context.Background(), UpdateDividendInput{
			ID:       "dividend-id",
			WalletID: "wallet-id",
			Year:     currentYear,
			Amount:   150000,
		})

		assert.ErrorIs(t, err, ErrDividendAlreadyExists)
	})

	t.Run("error - repository failure on lookup", func(t *testing.T) {
		repoErr := errors.New("db error")
		repo := &mockRepository{
			findByIDFunc: func(ctx context.Context, walletID string, id string) (*Dividend, error) {
				return nil, repoErr
			},
		}
		svc := NewService(repo)

		_, err := svc.Update(context.Background(), UpdateDividendInput{
			ID:       "dividend-id",
			WalletID: "wallet-id",
			Year:     currentYear,
			Amount:   150000,
		})

		assert.ErrorIs(t, err, repoErr)
	})

	t.Run("error - repository failure on update", func(t *testing.T) {
		repoErr := errors.New("db error")
		repo := &mockRepository{
			findByIDFunc: func(ctx context.Context, walletID string, id string) (*Dividend, error) {
				return &Dividend{ID: id, WalletID: "wallet-id", Year: 2024, Amount: 100000}, nil
			},
			updateFunc: func(ctx context.Context, dividend *Dividend) error {
				return repoErr
			},
		}
		svc := NewService(repo)

		_, err := svc.Update(context.Background(), UpdateDividendInput{
			ID:       "dividend-id",
			WalletID: "wallet-id",
			Year:     currentYear,
			Amount:   150000,
		})

		assert.ErrorIs(t, err, repoErr)
	})
}

func TestService_List(t *testing.T) {
	t.Run("success - returns dividends mapped to output", func(t *testing.T) {
		repo := &mockRepository{
			findByFilterFunc: func(ctx context.Context, filter DividendFilter) ([]Dividend, error) {
				return []Dividend{
					{ID: "dividend-2", WalletID: "wallet-id", Year: 2026, Amount: 200000},
					{ID: "dividend-1", WalletID: "wallet-id", Year: 2025, Amount: 150000},
				}, nil
			},
		}
		svc := NewService(repo)

		output, err := svc.List(context.Background(), DividendFilter{WalletID: "wallet-id"})

		require.NoError(t, err)
		require.Len(t, output, 2)
		assert.Equal(t, "dividend-2", output[0].ID)
		assert.Equal(t, 2026, output[0].Year)
		assert.Equal(t, int64(200000), output[0].Amount)
	})

	t.Run("success - returns empty list when no results", func(t *testing.T) {
		repo := &mockRepository{
			findByFilterFunc: func(ctx context.Context, filter DividendFilter) ([]Dividend, error) {
				return []Dividend{}, nil
			},
		}
		svc := NewService(repo)

		output, err := svc.List(context.Background(), DividendFilter{WalletID: "wallet-id", Year: 2030})

		require.NoError(t, err)
		assert.Empty(t, output)
	})

	t.Run("error - repository failure", func(t *testing.T) {
		repoErr := errors.New("db error")
		repo := &mockRepository{
			findByFilterFunc: func(ctx context.Context, filter DividendFilter) ([]Dividend, error) {
				return nil, repoErr
			},
		}
		svc := NewService(repo)

		_, err := svc.List(context.Background(), DividendFilter{WalletID: "wallet-id"})

		assert.ErrorIs(t, err, repoErr)
	})
}
