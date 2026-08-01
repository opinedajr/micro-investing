package wallet

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Service interface {
	Create(ctx context.Context, input CreateWalletInput) (*WalletOutput, error)
	List(ctx context.Context, userID string) ([]WalletOutput, error)
	Find(ctx context.Context, id string) (*WalletOutput, error)
	Update(ctx context.Context, id string, input UpdateWalletInput) (*WalletOutput, error)
}

type walletService struct {
	repo Repository
}

func NewService(repo Repository) *walletService {
	return &walletService{repo: repo}
}

func (s *walletService) Create(ctx context.Context, input CreateWalletInput) (*WalletOutput, error) {
	existing, err := s.repo.FindByNameAndUserID(ctx, input.Name, input.UserID)
	if err == nil && existing != nil {
		return nil, ErrWalletNameAlreadyExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
		return nil, err
	}

	wallet := &Wallet{
		UserID:      input.UserID,
		Name:        input.Name,
		Description: input.Description,
	}

	if err := s.repo.Create(ctx, wallet); err != nil {
		return nil, err
	}

	return &WalletOutput{
		ID:          wallet.ID,
		Name:        wallet.Name,
		Description: wallet.Description,
		CreatedAt:   wallet.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   wallet.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *walletService) List(ctx context.Context, userID string) ([]WalletOutput, error) {
	wallets, err := s.repo.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	outputs := make([]WalletOutput, len(wallets))
	for i, wallet := range wallets {
		outputs[i] = WalletOutput{
			ID:          wallet.ID,
			Name:        wallet.Name,
			Description: wallet.Description,
			CreatedAt:   wallet.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   wallet.UpdatedAt.Format(time.RFC3339),
		}
	}
	return outputs, nil
}

func (s *walletService) Find(ctx context.Context, id string) (*WalletOutput, error) {
	wallet, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &WalletOutput{
		ID:          wallet.ID,
		Name:        wallet.Name,
		Description: wallet.Description,
		CreatedAt:   wallet.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   wallet.UpdatedAt.Format(time.RFC3339),
	}, nil
}

func (s *walletService) Update(ctx context.Context, id string, input UpdateWalletInput) (*WalletOutput, error) {
	wallet, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if wallet.Name != input.Name {
		existing, err := s.repo.FindByNameAndUserID(ctx, input.Name, wallet.UserID)
		if err == nil && existing != nil {
			return nil, ErrWalletNameAlreadyExists
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) && err != nil {
			return nil, err
		}
	}

	wallet.Name = input.Name
	wallet.Description = input.Description

	if err := s.repo.Update(ctx, wallet); err != nil {
		return nil, err
	}

	return &WalletOutput{
		ID:          wallet.ID,
		Name:        wallet.Name,
		Description: wallet.Description,
		CreatedAt:   wallet.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   wallet.UpdatedAt.Format(time.RFC3339),
	}, nil
}
