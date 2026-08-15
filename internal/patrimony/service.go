package patrimony

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Service interface {
	Create(ctx context.Context, input CreatePatrimonyInput) (*PatrimonyOutput, error)
	Update(ctx context.Context, id string, input UpdatePatrimonyInput) (*PatrimonyOutput, error)
	List(ctx context.Context, filter PatrimonyFilter) ([]PatrimonyOutput, error)
	CreateAsset(ctx context.Context, input CreateAssetInput) (*AssetOutput, error)
	UpdateAsset(ctx context.Context, id string, input UpdateAssetInput) (*AssetOutput, error)
	DeleteAsset(ctx context.Context, id string) error
	ListAssets(ctx context.Context, filter AssetFilter) ([]AssetOutput, error)
}

type patrimonyService struct {
	repo      PatrimonyRepository
	assetRepo AssetRepository
}

func NewService(repo PatrimonyRepository, assetRepo AssetRepository) Service {
	return &patrimonyService{repo: repo, assetRepo: assetRepo}
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

func (s *patrimonyService) CreateAsset(ctx context.Context, input CreateAssetInput) (*AssetOutput, error) {
	if s.assetRepo == nil {
		return nil, fmt.Errorf("asset repository not configured")
	}

	parsedDate, err := time.Parse(time.RFC3339, input.Date)
	if err != nil {
		return nil, ErrInvalidAssetDate
	}

	if err := validateAssetInput(input.Type, parsedDate, input.Description, input.Amount); err != nil {
		return nil, err
	}

	asset := &Asset{
		WalletID:    input.WalletID,
		Type:        input.Type,
		Date:        parsedDate,
		Description: input.Description,
		Amount:      input.Amount,
	}

	if err := s.repo.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := s.assetRepo.Create(txCtx, asset); err != nil {
			return err
		}
		return s.recalcPatrimony(txCtx, asset.WalletID, asset.Type, asset.Date.Year(), int(asset.Date.Month()))
	}); err != nil {
		return nil, err
	}

	return toAssetOutput(asset), nil
}

func (s *patrimonyService) UpdateAsset(ctx context.Context, id string, input UpdateAssetInput) (*AssetOutput, error) {
	if s.assetRepo == nil {
		return nil, fmt.Errorf("asset repository not configured")
	}

	parsedDate, err := time.Parse(time.RFC3339, input.Date)
	if err != nil {
		return nil, ErrInvalidAssetDate
	}

	if err := validateAssetInput(input.Type, parsedDate, input.Description, input.Amount); err != nil {
		return nil, err
	}

	var updatedAsset *Asset
	var oldYear, oldMonth int
	var oldType AssetType

	if err := s.repo.RunInTransaction(ctx, func(txCtx context.Context) error {
		existing, err := s.assetRepo.FindByID(txCtx, id)
		if err != nil {
			return err
		}

		oldYear = existing.Date.Year()
		oldMonth = int(existing.Date.Month())
		oldType = existing.Type

		existing.Type = input.Type
		existing.Date = parsedDate
		existing.Description = input.Description
		existing.Amount = input.Amount

		if err := s.assetRepo.Update(txCtx, existing); err != nil {
			return err
		}

		updatedAsset = existing

		newYear := parsedDate.Year()
		newMonth := int(parsedDate.Month())

		if oldYear != newYear || oldMonth != newMonth || oldType != input.Type {
			if err := s.recalcPatrimony(txCtx, input.WalletID, oldType, oldYear, oldMonth); err != nil {
				return err
			}
		}

		return s.recalcPatrimony(txCtx, input.WalletID, input.Type, newYear, newMonth)
	}); err != nil {
		return nil, err
	}

	return toAssetOutput(updatedAsset), nil
}

func (s *patrimonyService) DeleteAsset(ctx context.Context, id string) error {
	if s.assetRepo == nil {
		return fmt.Errorf("asset repository not configured")
	}

	return s.repo.RunInTransaction(ctx, func(txCtx context.Context) error {
		asset, err := s.assetRepo.FindByID(txCtx, id)
		if err != nil {
			return err
		}

		if err := s.assetRepo.Delete(txCtx, id); err != nil {
			return err
		}

		return s.recalcPatrimony(txCtx, asset.WalletID, asset.Type, asset.Date.Year(), int(asset.Date.Month()))
	})
}

func (s *patrimonyService) ListAssets(ctx context.Context, filter AssetFilter) ([]AssetOutput, error) {
	if filter.StartDate != nil && filter.EndDate != nil && filter.StartDate.After(*filter.EndDate) {
		return nil, ErrInvalidDateRange
	}

	assets, err := s.assetRepo.FindByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}

	outputs := make([]AssetOutput, len(assets))
	for i, a := range assets {
		outputs[i] = *toAssetOutput(&a)
	}
	return outputs, nil
}

func (s *patrimonyService) recalcPatrimony(ctx context.Context, walletID string, assetType AssetType, year int, month int) error {
	total, err := s.assetRepo.SumByWalletTypeAndMonth(ctx, walletID, assetType, year, month)
	if err != nil {
		return err
	}

	existing, err := s.repo.FindByWalletYearMonthType(ctx, walletID, year, month, assetType)
	if err != nil && !errors.Is(err, ErrPatrimonyNotFound) {
		return err
	}

	if existing == nil {
		if total == 0 {
			return nil
		}
		return s.repo.Create(ctx, &Patrimony{
			WalletID: walletID,
			Year:     year,
			Month:    month,
			Type:     assetType,
			Amount:   total,
		})
	}

	existing.Amount = total
	return s.repo.Update(ctx, existing)
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

func validateAssetInput(assetType AssetType, date time.Time, description string, amount int64) error {
	if !assetType.IsValid() {
		return ErrInvalidAssetType
	}
	if date.After(time.Now()) {
		return ErrInvalidAssetDate
	}
	if len(description) < 3 || len(description) > 100 {
		return ErrInvalidAssetDescription
	}
	if amount <= 0 {
		return ErrInvalidAssetAmount
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

func toAssetOutput(a *Asset) *AssetOutput {
	return &AssetOutput{
		ID:          a.ID,
		WalletID:    a.WalletID,
		Type:        a.Type,
		Date:        a.Date.Format(time.RFC3339),
		Description: a.Description,
		Amount:      a.Amount,
		CreatedAt:   a.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   a.UpdatedAt.Format(time.RFC3339),
	}
}
