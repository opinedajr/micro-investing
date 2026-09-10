package position

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/opinedajr/micro-investing/internal/shared/logger"
	"github.com/opinedajr/micro-investing/internal/stock"
)

type Service interface {
	Create(ctx context.Context, input CreatePositionInput) (*PositionOutput, error)
	List(ctx context.Context, filter PositionFilter) ([]PositionOutput, error)
	Find(ctx context.Context, walletID string, id string) (*PositionOutput, error)
	ConsolidateByWallet(ctx context.Context, walletID string) error
}

type positionService struct {
	repo      Repository
	stockRepo stock.Repository
	validator *validator.Validate
	logger    logger.Logger
}

func NewService(repo Repository, stockRepo stock.Repository, logger logger.Logger) Service {
	return &positionService{
		repo:      repo,
		stockRepo: stockRepo,
		validator: validator.New(),
		logger:    logger,
	}
}

func (s *positionService) List(ctx context.Context, filter PositionFilter) ([]PositionOutput, error) {
	positions, err := s.repo.FindByFilter(ctx, filter)
	if err != nil {
		return nil, err
	}

	outputs := make([]PositionOutput, 0, len(positions))
	for i := range positions {
		outputs = append(outputs, *toPositionOutput(&positions[i]))
	}
	return outputs, nil
}

func (s *positionService) Find(ctx context.Context, walletID string, id string) (*PositionOutput, error) {
	position, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if position.WalletID != walletID {
		return nil, ErrPositionNotFound
	}

	return toPositionOutput(position), nil
}

func (s *positionService) Create(ctx context.Context, input CreatePositionInput) (*PositionOutput, error) {
	if err := s.validator.Struct(&input); err != nil {
		return nil, err
	}

	if _, err := s.stockRepo.FindByID(ctx, input.StockID); err != nil {
		if errors.Is(err, stock.ErrStockNotFound) {
			return nil, ErrStockNotFound
		}
		return nil, err
	}

	existing, err := s.repo.FindByWalletAndStockID(ctx, input.WalletID, input.StockID)
	if err != nil && !errors.Is(err, ErrPositionNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, ErrPositionAlreadyExists
	}

	position := &Position{
		WalletID:     input.WalletID,
		StockID:      input.StockID,
		Quantity:     input.Quantity,
		AveragePrice: input.AveragePrice,
	}

	var created *Position
	if err := s.repo.RunInTransaction(ctx, func(txCtx context.Context) error {
		if err := s.repo.Create(txCtx, position); err != nil {
			return err
		}
		if err := s.consolidateInTransaction(txCtx, input.WalletID); err != nil {
			return err
		}
		refreshed, err := s.repo.FindByID(txCtx, position.ID)
		if err != nil {
			return err
		}
		created = refreshed
		return nil
	}); err != nil {
		return nil, err
	}

	return toPositionOutput(created), nil
}

func (s *positionService) ConsolidateByWallet(ctx context.Context, walletID string) error {
	return s.repo.RunInTransaction(ctx, func(txCtx context.Context) error {
		return s.consolidateInTransaction(txCtx, walletID)
	})
}

func (s *positionService) consolidateInTransaction(ctx context.Context, walletID string) error {
	positions, err := s.repo.FindByFilter(ctx, PositionFilter{WalletID: walletID})
	if err != nil {
		return err
	}

	if len(positions) == 0 {
		return nil
	}

	stockIDs := make([]string, len(positions))
	for i, p := range positions {
		stockIDs[i] = p.StockID
	}

	priceMap, err := s.repo.FindCurrentPricesMap(ctx, stockIDs)
	if err != nil {
		return err
	}

	var totalInvested int64
	for i := range positions {
		price, ok := priceMap[positions[i].StockID]
		if !ok {
			price = 0
			s.logger.Warn(ctx, "current price not found for stock", "stock_id", positions[i].StockID)
		}
		positions[i].CurrentPrice = price
		positions[i].Invested = positions[i].Quantity * positions[i].AveragePrice
		positions[i].Balance = positions[i].Quantity * price
		positions[i].VariationValue = positions[i].Balance - positions[i].Invested
		if positions[i].AveragePrice != 0 {
			positions[i].VariationPercent = float64(positions[i].CurrentPrice-positions[i].AveragePrice) / float64(positions[i].AveragePrice) * 100
		} else {
			positions[i].VariationPercent = 0
		}
		totalInvested += positions[i].Invested
	}

	for i := range positions {
		if totalInvested > 0 {
			positions[i].PortfolioPercent = float64(positions[i].Invested) / float64(totalInvested) * 100
		} else {
			positions[i].PortfolioPercent = 0
		}
		if err := s.repo.Update(ctx, &positions[i]); err != nil {
			return err
		}
	}

	return nil
}

func toPositionOutput(p *Position) *PositionOutput {
	return &PositionOutput{
		ID:               p.ID,
		WalletID:         p.WalletID,
		StockID:          p.StockID,
		Quantity:         p.Quantity,
		AveragePrice:     p.AveragePrice,
		CurrentPrice:     p.CurrentPrice,
		Invested:         p.Invested,
		Balance:          p.Balance,
		VariationValue:   p.VariationValue,
		VariationPercent: p.VariationPercent,
		PortfolioPercent: p.PortfolioPercent,
		CreatedAt:        p.CreatedAt.Format(time.RFC3339),
		UpdatedAt:        p.UpdatedAt.Format(time.RFC3339),
	}
}
