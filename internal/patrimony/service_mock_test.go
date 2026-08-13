package patrimony

import "context"

type mockPatrimonyRepository struct {
	createFunc                  func(ctx context.Context, patrimony *Patrimony) error
	updateFunc                  func(ctx context.Context, patrimony *Patrimony) error
	findByIDFunc                func(ctx context.Context, id string) (*Patrimony, error)
	findByFilterFunc            func(ctx context.Context, filter PatrimonyFilter) ([]Patrimony, error)
	findByWalletYearMonthTypeFn func(ctx context.Context, walletID string, year int, month int, assetType AssetType) (*Patrimony, error)
	runInTransactionFn          func(ctx context.Context, fn func(ctx context.Context) error) error
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

func (m *mockPatrimonyRepository) RunInTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	if m.runInTransactionFn != nil {
		return m.runInTransactionFn(ctx, fn)
	}
	return fn(ctx)
}

type mockAssetRepository struct {
	createFunc       func(ctx context.Context, asset *Asset) error
	updateFunc       func(ctx context.Context, asset *Asset) error
	deleteFunc       func(ctx context.Context, id string) error
	findByIDFunc     func(ctx context.Context, id string) (*Asset, error)
	findByFilterFunc func(ctx context.Context, filter AssetFilter) ([]Asset, error)
	sumFunc          func(ctx context.Context, walletID string, assetType AssetType, year int, month int) (int64, error)
}

func (m *mockAssetRepository) Create(ctx context.Context, asset *Asset) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, asset)
	}
	return nil
}

func (m *mockAssetRepository) Update(ctx context.Context, asset *Asset) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, asset)
	}
	return nil
}

func (m *mockAssetRepository) Delete(ctx context.Context, id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func (m *mockAssetRepository) FindByID(ctx context.Context, id string) (*Asset, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, ErrAssetNotFound
}

func (m *mockAssetRepository) FindByFilter(ctx context.Context, filter AssetFilter) ([]Asset, error) {
	if m.findByFilterFunc != nil {
		return m.findByFilterFunc(ctx, filter)
	}
	return nil, nil
}

func (m *mockAssetRepository) SumByWalletTypeAndMonth(ctx context.Context, walletID string, assetType AssetType, year int, month int) (int64, error) {
	if m.sumFunc != nil {
		return m.sumFunc(ctx, walletID, assetType, year, month)
	}
	return 0, nil
}
