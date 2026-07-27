package wallet

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
)

type Service interface {
	Create(ctx context.Context, input CreateWalletInput) (*WalletOutput, error)
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
