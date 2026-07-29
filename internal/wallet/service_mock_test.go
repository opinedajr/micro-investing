package wallet

import (
	"context"
	"errors"
)

type mockService struct {
	createFunc func(ctx context.Context, input CreateWalletInput) (*WalletOutput, error)
	listFunc   func(ctx context.Context, userID string) ([]WalletOutput, error)
	findFunc   func(ctx context.Context, id string) (*WalletOutput, error)
}

func (m *mockService) Create(ctx context.Context, input CreateWalletInput) (*WalletOutput, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) List(ctx context.Context, userID string) ([]WalletOutput, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, userID)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) Find(ctx context.Context, id string) (*WalletOutput, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func newMockService(createFunc func(ctx context.Context, input CreateWalletInput) (*WalletOutput, error)) Service {
	return &mockService{createFunc: createFunc}
}

func newMockServiceWithListAndFind(
	createFunc func(ctx context.Context, input CreateWalletInput) (*WalletOutput, error),
	listFunc func(ctx context.Context, userID string) ([]WalletOutput, error),
	findFunc func(ctx context.Context, id string) (*WalletOutput, error),
) Service {
	return &mockService{
		createFunc: createFunc,
		listFunc:   listFunc,
		findFunc:   findFunc,
	}
}
