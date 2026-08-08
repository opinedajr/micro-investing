package patrimony

import (
	"context"
	"errors"
)

type mockService struct {
	createFunc func(ctx context.Context, input CreatePatrimonyInput) (*PatrimonyOutput, error)
	updateFunc func(ctx context.Context, id string, input UpdatePatrimonyInput) (*PatrimonyOutput, error)
	listFunc   func(ctx context.Context, filter PatrimonyListFilter) ([]PatrimonyOutput, error)
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

func (m *mockService) List(ctx context.Context, filter PatrimonyListFilter) ([]PatrimonyOutput, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, filter)
	}
	return nil, errors.New("not implemented")
}

func newMockService(createFunc func(ctx context.Context, input CreatePatrimonyInput) (*PatrimonyOutput, error)) Service {
	return &mockService{createFunc: createFunc}
}

func newMockServiceWithUpdate(updateFunc func(ctx context.Context, id string, input UpdatePatrimonyInput) (*PatrimonyOutput, error)) Service {
	return &mockService{updateFunc: updateFunc}
}

func newMockServiceWithList(listFunc func(ctx context.Context, filter PatrimonyListFilter) ([]PatrimonyOutput, error)) Service {
	return &mockService{listFunc: listFunc}
}
