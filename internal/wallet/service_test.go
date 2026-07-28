package wallet

import (
	"context"
	"errors"
	"time"

	"github.com/opinedajr/micro-investing/internal/shared"
	"github.com/stretchr/testify/assert"
	"testing"

	"gorm.io/gorm"
)

type mockRepository struct {
	createFunc          func(ctx context.Context, wallet *Wallet) error
	findByNameAndUserID func(ctx context.Context, name string, userID string) (*Wallet, error)
	findByIDFunc        func(ctx context.Context, id string) (*Wallet, error)
	findAllByUserIDFunc func(ctx context.Context, userID string) ([]Wallet, error)
}

func (m *mockRepository) Create(ctx context.Context, wallet *Wallet) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, wallet)
	}
	return nil
}

func (m *mockRepository) FindAllByUserID(ctx context.Context, userID string) ([]Wallet, error) {
	if m.findAllByUserIDFunc != nil {
		return m.findAllByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockRepository) FindByID(ctx context.Context, id string) (*Wallet, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockRepository) FindByNameAndUserID(ctx context.Context, name string, userID string) (*Wallet, error) {
	if m.findByNameAndUserID != nil {
		return m.findByNameAndUserID(ctx, name, userID)
	}
	return nil, errors.New("not found")
}

func (m *mockRepository) Update(ctx context.Context, wallet *Wallet) error {
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, id string) error {
	return nil
}

func TestService_Create(t *testing.T) {
	t.Run("success - creates wallet and returns output", func(t *testing.T) {
		now := time.Now()
		mockRepo := &mockRepository{
			createFunc: func(ctx context.Context, wallet *Wallet) error {
				wallet.ID = "generated-uuid"
				wallet.CreatedAt = now
				wallet.UpdatedAt = now
				return nil
			},
			findByNameAndUserID: func(ctx context.Context, name string, userID string) (*Wallet, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}

		service := NewService(mockRepo)

		input := CreateWalletInput{
			UserID:      shared.DefaultUserID,
			Name:        "Minha Carteira",
			Description: ptrStr("Ações e FIIs"),
		}

		output, err := service.Create(context.Background(), input)
		if err != nil {
			t.Logf("Error: %v", err)
		}
		assert.NoError(t, err)
		assert.NotEmpty(t, output.ID)
		assert.Equal(t, "Minha Carteira", output.Name)
		assert.Equal(t, "Ações e FIIs", *output.Description)
	})

	t.Run("error - returns error when name already exists", func(t *testing.T) {
		existingWallet := &Wallet{
			ID:     "existing-id",
			UserID: shared.DefaultUserID,
			Name:   "Duplicada",
		}

		mockRepo := &mockRepository{
			createFunc: func(ctx context.Context, wallet *Wallet) error {
				return nil
			},
			findByNameAndUserID: func(ctx context.Context, name string, userID string) (*Wallet, error) {
				return existingWallet, nil
			},
		}

		service := NewService(mockRepo)

		input := CreateWalletInput{
			UserID:      shared.DefaultUserID,
			Name:        "Duplicada",
			Description: ptrStr("Descrição"),
		}

		_, err := service.Create(context.Background(), input)
		assert.Error(t, err)
		assert.Equal(t, ErrWalletNameAlreadyExists, err)
	})
}

func ptrStr(s string) *string {
	return &s
}

func TestService_List(t *testing.T) {
	t.Run("success - lists wallets for user", func(t *testing.T) {
		now := time.Now()
		mockRepo := &mockRepository{
			findAllByUserIDFunc: func(ctx context.Context, userID string) ([]Wallet, error) {
				return []Wallet{
					{
						ID:          "wallet-1",
						UserID:      userID,
						Name:        "Carteira 1",
						Description: ptrStr("Desc 1"),
						CreatedAt:   now,
						UpdatedAt:   now,
					},
					{
						ID:          "wallet-2",
						UserID:      userID,
						Name:        "Carteira 2",
						Description: ptrStr("Desc 2"),
						CreatedAt:   now,
						UpdatedAt:   now,
					},
				}, nil
			},
		}

		service := NewService(mockRepo)

		outputs, err := service.List(context.Background(), shared.DefaultUserID)
		assert.NoError(t, err)
		assert.Len(t, outputs, 2)
		assert.Equal(t, "wallet-1", outputs[0].ID)
		assert.Equal(t, "Carteira 1", outputs[0].Name)
		assert.Equal(t, "wallet-2", outputs[1].ID)
		assert.Equal(t, "Carteira 2", outputs[1].Name)
	})
}

func TestService_Find(t *testing.T) {
	t.Run("success - finds wallet by id", func(t *testing.T) {
		now := time.Now()
		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, id string) (*Wallet, error) {
				return &Wallet{
					ID:          id,
					UserID:      shared.DefaultUserID,
					Name:        "Carteira Encontrada",
					Description: ptrStr("Descrição"),
					CreatedAt:   now,
					UpdatedAt:   now,
				}, nil
			},
		}

		service := NewService(mockRepo)

		output, err := service.Find(context.Background(), "wallet-123")
		assert.NoError(t, err)
		assert.Equal(t, "wallet-123", output.ID)
		assert.Equal(t, "Carteira Encontrada", output.Name)
		assert.Equal(t, "Descrição", *output.Description)
	})

	t.Run("error - returns error when wallet not found", func(t *testing.T) {
		mockRepo := &mockRepository{
			findByIDFunc: func(ctx context.Context, id string) (*Wallet, error) {
				return nil, gorm.ErrRecordNotFound
			},
		}

		service := NewService(mockRepo)

		_, err := service.Find(context.Background(), "non-existent")
		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
	})
}
