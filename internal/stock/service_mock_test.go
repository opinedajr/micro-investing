package stock

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRepository struct {
	findByTickerFunc func(ctx context.Context, ticker string) (*Stock, error)
	listFunc         func(ctx context.Context) ([]Stock, error)
	seedFunc         func(ctx context.Context, stocks []Stock, force bool) error
}

func (m *mockRepository) Create(ctx context.Context, stock *Stock) error {
	return errors.New("not implemented")
}

func (m *mockRepository) FindByTicker(ctx context.Context, ticker string) (*Stock, error) {
	if m.findByTickerFunc != nil {
		return m.findByTickerFunc(ctx, ticker)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) FindByID(ctx context.Context, id string) (*Stock, error) {
	return nil, errors.New("not implemented")
}

func (m *mockRepository) List(ctx context.Context) ([]Stock, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx)
	}
	return nil, errors.New("not implemented")
}

func (m *mockRepository) Seed(ctx context.Context, stocks []Stock, force bool) error {
	if m.seedFunc != nil {
		return m.seedFunc(ctx, stocks, force)
	}
	return errors.New("not implemented")
}

func newMockRepositoryWithFind(findByTickerFunc func(ctx context.Context, ticker string) (*Stock, error)) Repository {
	return &mockRepository{findByTickerFunc: findByTickerFunc}
}

func newMockRepositoryWithList(listFunc func(ctx context.Context) ([]Stock, error)) Repository {
	return &mockRepository{listFunc: listFunc}
}

func newMockRepositoryWithSeed(seedFunc func(ctx context.Context, stocks []Stock, force bool) error) Repository {
	return &mockRepository{seedFunc: seedFunc}
}

func TestService_FindByTicker(t *testing.T) {
	t.Run("success - finds stock by ticker", func(t *testing.T) {
		repo := newMockRepositoryWithFind(func(ctx context.Context, ticker string) (*Stock, error) {
			return &Stock{
				ID:     "stock-id",
				Ticker: ticker,
				Name:   "Petrobras PN",
				Sector: "Petróleo",
				Rank:   10,
			}, nil
		})

		service := NewService(repo)
		output, err := service.FindByTicker(context.Background(), "PETR4")

		require.NoError(t, err)
		assert.Equal(t, "stock-id", output.ID)
		assert.Equal(t, "PETR4", output.Ticker)
		assert.Equal(t, "Petrobras PN", output.Name)
	})

	t.Run("error - returns not found when stock does not exist", func(t *testing.T) {
		repo := newMockRepositoryWithFind(func(ctx context.Context, ticker string) (*Stock, error) {
			return nil, ErrStockNotFound
		})

		service := NewService(repo)
		output, err := service.FindByTicker(context.Background(), "INVALID")

		assert.ErrorIs(t, err, ErrStockNotFound)
		assert.Nil(t, output)
	})
}

func TestService_List(t *testing.T) {
	t.Run("success - lists stocks ordered", func(t *testing.T) {
		repo := newMockRepositoryWithList(func(ctx context.Context) ([]Stock, error) {
			return []Stock{
				{ID: "1", Ticker: "B3SA3", Name: "B3", Sector: "Bolsa", Rank: 9},
				{ID: "2", Ticker: "PETR4", Name: "Petrobras", Sector: "Petróleo", Rank: 10},
			}, nil
		})

		service := NewService(repo)
		outputs, err := service.List(context.Background())

		require.NoError(t, err)
		assert.Len(t, outputs, 2)
		assert.Equal(t, "B3SA3", outputs[0].Ticker)
		assert.Equal(t, "PETR4", outputs[1].Ticker)
	})

	t.Run("error - returns repository error", func(t *testing.T) {
		repo := newMockRepositoryWithList(func(ctx context.Context) ([]Stock, error) {
			return nil, errors.New("database error")
		})

		service := NewService(repo)
		outputs, err := service.List(context.Background())

		assert.Error(t, err)
		assert.Nil(t, outputs)
	})
}

func TestService_Seed(t *testing.T) {
	t.Run("success - seeds blue chips idempotently", func(t *testing.T) {
		repo := newMockRepositoryWithSeed(func(ctx context.Context, stocks []Stock, force bool) error {
			assert.Len(t, stocks, len(BlueChips()))
			assert.False(t, force)
			return nil
		})

		service := NewService(repo)
		err := service.Seed(context.Background(), SeedInput{Force: false})

		assert.NoError(t, err)
	})

	t.Run("success - seeds with force", func(t *testing.T) {
		repo := newMockRepositoryWithSeed(func(ctx context.Context, stocks []Stock, force bool) error {
			assert.True(t, force)
			return nil
		})

		service := NewService(repo)
		err := service.Seed(context.Background(), SeedInput{Force: true})

		assert.NoError(t, err)
	})

	t.Run("error - returns validation error for invalid seed data", func(t *testing.T) {
		repo := newMockRepositoryWithSeed(func(ctx context.Context, stocks []Stock, force bool) error {
			return nil
		})

		service := NewService(repo)
		original := BlueChips
		BlueChips = func() []Stock {
			return []Stock{{Ticker: "AB", Name: "A", Sector: "B", Rank: 0}}
		}
		defer func() { BlueChips = original }()

		err := service.Seed(context.Background(), SeedInput{})

		assert.Error(t, err)
	})
}
