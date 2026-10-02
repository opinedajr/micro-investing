package dividends

import "context"

type serviceFuncs struct {
	createFunc func(ctx context.Context, input CreateDividendInput) (*DividendOutput, error)
	listFunc   func(ctx context.Context, filter DividendFilter) ([]DividendOutput, error)
}

type mockService struct {
	serviceFuncs
}

func newMockService(fn serviceFuncs) *mockService {
	return &mockService{serviceFuncs: fn}
}

func (m *mockService) Create(ctx context.Context, input CreateDividendInput) (*DividendOutput, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, input)
	}
	return nil, nil
}

func (m *mockService) List(ctx context.Context, filter DividendFilter) ([]DividendOutput, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return nil, nil
}
