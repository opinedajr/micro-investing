package wallet

import (
	"context"
	"errors"
)

type mockService struct {
	createFunc func(ctx context.Context, input CreateWalletInput) (*WalletOutput, error)
}

func (m *mockService) Create(ctx context.Context, input CreateWalletInput) (*WalletOutput, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}

func newMockService(createFunc func(ctx context.Context, input CreateWalletInput) (*WalletOutput, error)) Service {
	return &mockService{createFunc: createFunc}
}
