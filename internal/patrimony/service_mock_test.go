package patrimony

import "context"

type mockPatrimonyRepository struct {
	createFunc                  func(ctx context.Context, patrimony *Patrimony) error
	updateFunc                  func(ctx context.Context, patrimony *Patrimony) error
	findByIDFunc                func(ctx context.Context, id string) (*Patrimony, error)
	findByFilterFunc            func(ctx context.Context, filter PatrimonyFilter) ([]Patrimony, error)
	findByWalletYearMonthTypeFn func(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error)
}

func (m *mockPatrimonyRepository) Create(ctx context.Context, patrimony *Patrimony) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, patrimony)
	}
	return nil
}

func (m *mockPatrimonyRepository) Update(ctx context.Context, patrimony *Patrimony) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, patrimony)
	}
	return nil
}

func (m *mockPatrimonyRepository) FindByID(ctx context.Context, id string) (*Patrimony, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, ErrPatrimonyNotFound
}

func (m *mockPatrimonyRepository) FindByFilter(ctx context.Context, filter PatrimonyFilter) ([]Patrimony, error) {
	if m.findByFilterFunc != nil {
		return m.findByFilterFunc(ctx, filter)
	}
	return nil, nil
}

func (m *mockPatrimonyRepository) FindByWalletYearMonthType(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error) {
	if m.findByWalletYearMonthTypeFn != nil {
		return m.findByWalletYearMonthTypeFn(ctx, walletID, year, month, assetType)
	}
	return nil, ErrPatrimonyNotFound
}
