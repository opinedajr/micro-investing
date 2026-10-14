package dividends

import (
	"context"
	"errors"
	"time"
)

type Service interface {
	Create(ctx context.Context, input CreateDividendInput) (*DividendOutput, error)
	Update(ctx context.Context, input UpdateDividendInput) (*DividendOutput, error)
	List(ctx context.Context, filter DividendFilter) ([]DividendOutput, error)
}

type dividendService struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &dividendService{repo: repo}
}

func (s *dividendService) Create(ctx context.Context, input CreateDividendInput) (*DividendOutput, error) {
	if err := validateDividendInput(input.Year, input.Amount); err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByWalletYear(ctx, input.WalletID, input.Year)
	if err != nil && !errors.Is(err, ErrDividendNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrDividendAlreadyExists
	}

	dividend := &Dividend{
		WalletID: input.WalletID,
		Year:     input.Year,
		Amount:   input.Amount,
	}

	if err := s.repo.Create(ctx, dividend); err != nil {
		return nil, err
	}

	return toOutput(dividend), nil
}

func (s *dividendService) Update(ctx context.Context, input UpdateDividendInput) (*DividendOutput, error) {
	if err := validateDividendInput(input.Year, input.Amount); err != nil {
		return nil, err
	}

	dividend, err := s.repo.FindByID(ctx, input.WalletID, input.ID)
	if err != nil {
		return nil, err
	}

	if dividend.Year != input.Year {
		existing, err := s.repo.FindByWalletYear(ctx, input.WalletID, input.Year)
		if err != nil && !errors.Is(err, ErrDividendNotFound) {
			return nil, err
		}
		if existing != nil && existing.ID != dividend.ID {
			return nil, ErrDividendAlreadyExists
		}
	}

	dividend.Year = input.Year
	dividend.Amount = input.Amount

	if err := s.repo.Update(ctx, dividend); err != nil {
		return nil, err
	}

	return toOutput(dividend), nil
}

func (s *dividendService) List(ctx context.Context, filter DividendFilter) ([]DividendOutput, error) {
	dividends, err := s.repo.FindByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}

	outputs := make([]DividendOutput, len(dividends))
	for i, d := range dividends {
		outputs[i] = *toOutput(&d)
	}
	return outputs, nil
}

func validateDividendInput(year int, amount int64) error {
	currentYear := time.Now().Year()
	if year < 1900 || year > currentYear+1 {
		return ErrInvalidDividendYear
	}
	if amount <= 0 {
		return ErrInvalidDividendAmount
	}
	return nil
}

func toOutput(d *Dividend) *DividendOutput {
	return &DividendOutput{
		ID:     d.ID,
		Year:   d.Year,
		Amount: d.Amount,
	}
}
