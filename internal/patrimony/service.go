package patrimony

import (
	"context"
	"errors"
	"time"
)

type Service interface {
	Create(ctx context.Context, input CreatePatrimonyInput) (*PatrimonyOutput, error)
	Update(ctx context.Context, id string, input UpdatePatrimonyInput) (*PatrimonyOutput, error)
	List(ctx context.Context, filter PatrimonyFilter) ([]PatrimonyOutput, error)
}

type patrimonyService struct {
	repo PatrimonyRepository
}

func NewService(repo PatrimonyRepository) Service {
	return &patrimonyService{repo: repo}
}

func (s *patrimonyService) Create(ctx context.Context, input CreatePatrimonyInput) (*PatrimonyOutput, error) {
	if err := validatePatrimonyInput(input.Year, input.Month, input.Type, input.Amount); err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByWalletYearMonthType(ctx, input.WalletID, input.Year, input.Month, input.Type)
	if err == nil && existing != nil {
		return nil, ErrPatrimonyAlreadyExists
	}
	if !errors.Is(err, ErrPatrimonyNotFound) {
		return nil, err
	}

	patrimony := &Patrimony{
		WalletID: input.WalletID,
		Year:     input.Year,
		Month:    input.Month,
		Type:     input.Type,
		Amount:   input.Amount,
	}

	if err := s.repo.Create(ctx, patrimony); err != nil {
		return nil, err
	}

	return toOutput(patrimony), nil
}

func (s *patrimonyService) Update(ctx context.Context, id string, input UpdatePatrimonyInput) (*PatrimonyOutput, error) {
	if err := validatePatrimonyInput(input.Year, input.Month, input.Type, input.Amount); err != nil {
		return nil, err
	}

	patrimony, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if patrimony.WalletID != input.WalletID {
		return nil, ErrPatrimonyNotFound
	}

	if patrimony.Year != input.Year || patrimony.Month != input.Month || patrimony.Type != input.Type {
		existing, err := s.repo.FindByWalletYearMonthType(ctx, input.WalletID, input.Year, input.Month, input.Type)
		if err == nil && existing != nil && existing.ID != id {
			return nil, ErrPatrimonyAlreadyExists
		}
		if !errors.Is(err, ErrPatrimonyNotFound) {
			return nil, err
		}
	}

	patrimony.Year = input.Year
	patrimony.Month = input.Month
	patrimony.Type = input.Type
	patrimony.Amount = input.Amount

	if err := s.repo.Update(ctx, patrimony); err != nil {
		return nil, err
	}

	return toOutput(patrimony), nil
}

func (s *patrimonyService) List(ctx context.Context, filter PatrimonyFilter) ([]PatrimonyOutput, error) {
	patrimonies, err := s.repo.FindByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}

	outputs := make([]PatrimonyOutput, len(patrimonies))
	for i, p := range patrimonies {
		outputs[i] = *toOutput(&p)
	}
	return outputs, nil
}

func validatePatrimonyInput(year int, month int, assetType AssetType, amount int64) error {
	currentYear := time.Now().Year()
	if year < 2000 || year > currentYear+1 {
		return ErrInvalidPatrimonyYear
	}
	if month < 1 || month > 12 {
		return ErrInvalidPatrimonyMonth
	}
	if !assetType.IsValid() {
		return ErrInvalidAssetType
	}
	if amount < 0 {
		return ErrInvalidPatrimonyAmount
	}
	return nil
}

func toOutput(p *Patrimony) *PatrimonyOutput {
	return &PatrimonyOutput{
		ID:        p.ID,
		WalletID:  p.WalletID,
		Year:      p.Year,
		Month:     p.Month,
		Type:      p.Type,
		Amount:    p.Amount,
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
	}
}
