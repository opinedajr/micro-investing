package dashboard

import (
	"context"

	"github.com/opinedajr/micro-investing/internal/patrimony"
	"github.com/opinedajr/micro-investing/internal/position"
	"github.com/opinedajr/micro-investing/internal/stock"
)

type Service interface {
	Summary(ctx context.Context, walletID string) (*SummaryOutput, error)
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
