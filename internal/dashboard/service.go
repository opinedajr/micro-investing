package dashboard

import (
	"context"

	"github.com/opinedajr/micro-investing/internal/patrimony"
	"github.com/opinedajr/micro-investing/internal/position"
	"github.com/opinedajr/micro-investing/internal/stock"
)

type Service interface {
	Summary(ctx context.Context, walletID string) (*SummaryOutput, error)
	Allocation(ctx context.Context, walletID string) (*AllocationOutput, error)
}

type dashboardService struct {
	patrimonyRepository patrimony.PatrimonyRepository
	positionRepository  position.Repository
	stockRepository     stock.Repository
}

func NewService(patrimonyRepository patrimony.PatrimonyRepository, positionRepository position.Repository, stockRepository stock.Repository) Service {
	return &dashboardService{
		patrimonyRepository: patrimonyRepository,
		positionRepository:  positionRepository,
		stockRepository:     stockRepository,
	}
}

func (s *dashboardService) Allocation(ctx context.Context, walletID string) (*AllocationOutput, error) {
	year, month, err := s.patrimonyRepository.FindLatestMonthByWallet(ctx, walletID)
	if err != nil {
		return nil, err
	}

	if year == 0 || month == 0 {
		return &AllocationOutput{
			Items: []AllocationItem{},
			Total: 0,
		}, nil
	}

	typeAmounts, err := s.patrimonyRepository.SumByWalletYearMonth(ctx, walletID, year, month)
	if err != nil {
		return nil, err
	}

	var total int64
	for _, item := range typeAmounts {
		if item.Amount > 0 {
			total += item.Amount
		}
	}

	items := make([]AllocationItem, 0)
	for _, item := range typeAmounts {
		if item.Amount <= 0 {
			continue
		}
		percentage := 0.0
		if total > 0 {
			percentage = float64(item.Amount) / float64(total) * 100
		}
		items = append(items, AllocationItem{
			Type:       string(item.Type),
			Amount:     item.Amount,
			Percentage: percentage,
		})
	}

	return &AllocationOutput{
		Items: items,
		Total: total,
	}, nil
}

func (s *dashboardService) Summary(ctx context.Context, walletID string) (*SummaryOutput, error) {
	year, month, err := s.patrimonyRepository.FindLatestMonthByWallet(ctx, walletID)
	if err != nil {
		return nil, err
	}

	var currentPatrimony int64
	if year > 0 && month > 0 {
		typeAmounts, err := s.patrimonyRepository.SumByWalletYearMonth(ctx, walletID, year, month)
		if err != nil {
			return nil, err
		}
		for _, item := range typeAmounts {
			currentPatrimony += item.Amount
		}
	}

	stocksInvested, err := s.positionRepository.SumInvestedByWallet(ctx, walletID)
	if err != nil {
		return nil, err
	}

	return &SummaryOutput{
		CurrentPatrimony: currentPatrimony,
		YearlyDividends:  0,
		StocksInvested:   stocksInvested,
	}, nil
}
