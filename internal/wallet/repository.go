package wallet

import "context"

type Repository interface {
	Create(ctx context.Context, wallet *Wallet) error
	FindAllByUserID(ctx context.Context, userID string) ([]Wallet, error)
	FindByID(ctx context.Context, id string) (*Wallet, error)
	FindByNameAndUserID(ctx context.Context, name string, userID string) (*Wallet, error)
	Update(ctx context.Context, wallet *Wallet) error
	Delete(ctx context.Context, id string) error
}
