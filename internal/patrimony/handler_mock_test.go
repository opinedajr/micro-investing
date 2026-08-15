package patrimony

import (
	"context"
	"errors"
)

type mockService struct {
	createFunc      func(ctx context.Context, input CreatePatrimonyInput) (*PatrimonyOutput, error)
	updateFunc      func(ctx context.Context, id string, input UpdatePatrimonyInput) (*PatrimonyOutput, error)
	listFunc        func(ctx context.Context, filter PatrimonyFilter) ([]PatrimonyOutput, error)
	createAssetFunc func(ctx context.Context, input CreateAssetInput) (*AssetOutput, error)
	updateAssetFunc func(ctx context.Context, id string, input UpdateAssetInput) (*AssetOutput, error)
	deleteAssetFunc func(ctx context.Context, id string) error
	listAssetsFunc  func(ctx context.Context, filter AssetFilter) ([]AssetOutput, error)
}

func (m *mockService) Create(ctx context.Context, input CreatePatrimonyInput) (*PatrimonyOutput, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) Update(ctx context.Context, id string, input UpdatePatrimonyInput) (*PatrimonyOutput, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) List(ctx context.Context, filter PatrimonyFilter) ([]PatrimonyOutput, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) CreateAsset(ctx context.Context, input CreateAssetInput) (*AssetOutput, error) {
	if m.createAssetFunc != nil {
		return m.createAssetFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) UpdateAsset(ctx context.Context, id string, input UpdateAssetInput) (*AssetOutput, error) {
	if m.updateAssetFunc != nil {
		return m.updateAssetFunc(ctx, id, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockService) DeleteAsset(ctx context.Context, id string) error {
	if m.deleteAssetFunc != nil {
		return m.deleteAssetFunc(ctx, id)
	}
	return errors.New("not implemented")
}

func (m *mockService) ListAssets(ctx context.Context, filter AssetFilter) ([]AssetOutput, error) {
	if m.listAssetsFunc != nil {
		return m.listAssetsFunc(ctx, filter)
	}
	return nil, errors.New("not implemented")
}

func newMockService(createFunc func(ctx context.Context, input CreatePatrimonyInput) (*PatrimonyOutput, error)) Service {
	return &mockService{createFunc: createFunc}
}

func newMockServiceWithUpdate(updateFunc func(ctx context.Context, id string, input UpdatePatrimonyInput) (*PatrimonyOutput, error)) Service {
	return &mockService{updateFunc: updateFunc}
}

func newMockServiceWithList(listFunc func(ctx context.Context, filter PatrimonyFilter) ([]PatrimonyOutput, error)) Service {
	return &mockService{listFunc: listFunc}
}
