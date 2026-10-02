package dividends

import "context"

type mockRepository struct {
	createFunc              func(ctx context.Context, dividend *Dividend) error
	findByFilterFunc        func(ctx context.Context, filter DividendFilter) ([]Dividend, error)
	findByWalletYearFunc    func(ctx context.Context, walletID string, year int) (*Dividend, error)
	findByIDFunc            func(ctx context.Context, walletID string, id string) (*Dividend, error)
	updateFunc              func(ctx context.Context, dividend *Dividend) error
	runInTransactionFunc    func(ctx context.Context, fn func(ctx context.Context) error) error
}

func (m *mockRepository) Create(ctx context.Context, dividend *Dividend) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, dividend)
	}
	return nil
}

func (m *mockRepository) FindByFilter(ctx context.Context, filter DividendFilter) ([]Dividend, error) {
	if m.findByFilterFunc != nil {
		return m.findByFilterFunc(ctx, filter)
	}
	return nil, nil
}

func (m *mockRepository) FindByWalletYear(ctx context.Context, walletID string, year int) (*Dividend, error) {
	if m.findByWalletYearFunc != nil {
		return m.findByWalletYearFunc(ctx, walletID, year)
	}
	return nil, ErrDividendNotFound
}

func (m *mockRepository) FindByID(ctx context.Context, walletID string, id string) (*Dividend, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, walletID, id)
	}
	return nil, ErrDividendNotFound
}

func (m *mockRepository) Update(ctx context.Context, dividend *Dividend) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, dividend)
	}
	return nil
}

func (m *mockRepository) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.runInTransactionFunc != nil {
		return m.runInTransactionFunc(ctx, fn)
	}
	return fn(ctx)
}
