package stock

import (
	"context"
	"regexp"
	"time"

	"github.com/go-playground/validator/v10"
)

type Service interface {
	FindByTicker(ctx context.Context, ticker string) (*StockOutput, error)
	List(ctx context.Context) ([]StockOutput, error)
	Seed(ctx context.Context, input SeedInput) error
}

type stockService struct {
	repo      Repository
	validator *validator.Validate
}

func NewService(repo Repository) Service {
	v := validator.New()
	_ = v.RegisterValidation("ticker", validateTicker)

	return &stockService{
		repo:      repo,
		validator: v,
	}
}

func (s *stockService) FindByTicker(ctx context.Context, ticker string) (*StockOutput, error) {
	stock, err := s.repo.FindByTicker(ctx, ticker)
	if err != nil {
		return nil, err
	}

	return toStockOutput(stock), nil
}

func (s *stockService) List(ctx context.Context) ([]StockOutput, error) {
	stocks, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	outputs := make([]StockOutput, len(stocks))
	for i, stock := range stocks {
		outputs[i] = *toStockOutput(&stock)
	}
	return outputs, nil
}

func (s *stockService) Seed(ctx context.Context, input SeedInput) error {
	stocks := BlueChips()

	for i := range stocks {
		if err := s.validator.Struct(&stocks[i]); err != nil {
			return err
		}
		if stocks[i].ID == "" {
			_ = stocks[i].BeforeCreate(nil)
		}
	}

	return s.repo.Seed(ctx, stocks, input.Force)
}

func toStockOutput(stock *Stock) *StockOutput {
	return &StockOutput{
		ID:        stock.ID,
		Ticker:    stock.Ticker,
		Name:      stock.Name,
		Sector:    stock.Sector,
		Rank:      stock.Rank,
		Website:   stock.Website,
		CreatedAt: stock.CreatedAt.Format(time.RFC3339),
		UpdatedAt: stock.UpdatedAt.Format(time.RFC3339),
	}
}

func validateTicker(fl validator.FieldLevel) bool {
	matched, _ := regexp.MatchString(`^[A-Z0-9]{4,6}$`, fl.Field().String())
	return matched
}
