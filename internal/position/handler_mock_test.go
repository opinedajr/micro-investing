package position

import (
	"context"
	"errors"
)

type mockService struct {
	createFunc           func(ctx context.Context, input CreatePositionInput) (*PositionOutput, error)
	consolidateByWalletFunc func(ctx context.Context, walletID string) error
}

func (m *mockService) Create(ctx context.Context, input CreatePositionInput) (*PositionOutput, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) ConsolidateByWallet(ctx context.Context, walletID string) error {
	if m.consolidateByWalletFunc != nil {
		return m.consolidateByWalletFunc(ctx, walletID)
	}
	return errors.New("not implemented")
}

func newMockServiceWithCreate(createFunc func(ctx context.Context, input CreatePositionInput) (*PositionOutput, error)) Service {
	return &mockService{createFunc: createFunc}
}
