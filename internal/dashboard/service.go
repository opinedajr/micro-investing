package dashboard

import (
	"context"
	"time"

	"github.com/opinedajr/micro-investing/internal/patrimony"
	"github.com/opinedajr/micro-investing/internal/position"
	"github.com/opinedajr/micro-investing/internal/stock"
)

type Service interface {
	Summary(ctx context.Context, walletID string) (*SummaryOutput, error)
	Allocation(ctx context.Context, walletID string) (*AllocationOutput, error)
	Evolution(ctx context.Context, walletID string, input EvolutionInput) (*EvolutionOutput, error)
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

func (s *dashboardService) Evolution(ctx context.Context, walletID string, input EvolutionInput) (*EvolutionOutput, error) {
	startYear, startMonth, endYear, endMonth := resolveEvolutionPeriod(time.Now(), input)

	totalSeries := make([]EvolutionMonthOutput, 0)
	fixedIncomeSeries := make([]EvolutionMonthOutput, 0)
	stocksSeries := make([]EvolutionMonthOutput, 0)
	emergencyReserveSeries := make([]EvolutionMonthOutput, 0)

	var previousTotal int64
	var previousFixedIncome int64
	var previousStocks int64
	var previousEmergencyReserve int64

	year := startYear
	month := startMonth

	for {
		typeAmounts, err := s.patrimonyRepository.SumByWalletYearMonth(ctx, walletID, year, month)
		if err != nil {
			return nil, err
		}

		var monthTotal int64
		var monthFixedIncome int64
		var monthStocks int64
		var monthEmergencyReserve int64

		if len(typeAmounts) > 0 {
			for _, item := range typeAmounts {
				monthTotal += item.Amount
				if item.Type == patrimony.TypeFixedIncome {
					monthFixedIncome += item.Amount
				}
				if item.Type == patrimony.TypeStocks {
					monthStocks += item.Amount
				}
				if item.Type == patrimony.TypeEmergencyReserve {
					monthEmergencyReserve += item.Amount
				}
			}
			previousTotal = monthTotal
			previousFixedIncome = monthFixedIncome
			previousStocks = monthStocks
			previousEmergencyReserve = monthEmergencyReserve
		} else {
			monthTotal = previousTotal
			monthFixedIncome = previousFixedIncome
			monthStocks = previousStocks
			monthEmergencyReserve = previousEmergencyReserve
		}

		totalSeries = append(totalSeries, EvolutionMonthOutput{Year: year, Month: month, Amount: monthTotal})
		fixedIncomeSeries = append(fixedIncomeSeries, EvolutionMonthOutput{Year: year, Month: month, Amount: monthFixedIncome})
		stocksSeries = append(stocksSeries, EvolutionMonthOutput{Year: year, Month: month, Amount: monthStocks})
		emergencyReserveSeries = append(emergencyReserveSeries, EvolutionMonthOutput{Year: year, Month: month, Amount: monthEmergencyReserve})

		if year == endYear && month == endMonth {
			break
		}

		month++
		if month > 12 {
			month = 1
			year++
		}
	}

	return &EvolutionOutput{
		Total: totalSeries,
		ByCategory: EvolutionCategoryOutput{
			FixedIncome:      fixedIncomeSeries,
			Stocks:           stocksSeries,
			EmergencyReserve: emergencyReserveSeries,
		},
	}, nil
}

func resolveEvolutionPeriod(reference time.Time, input EvolutionInput) (startYear int, startMonth int, endYear int, endMonth int) {
	if input.Year > 0 {
		endYear = input.Year
		if input.Quarter > 0 {
			startMonth = (input.Quarter-1)*3 + 1
			endMonth = input.Quarter * 3
			startYear = input.Year
			return startYear, startMonth, endYear, endMonth
		}
		startYear = input.Year
		startMonth = 1
		endMonth = 12
		return startYear, startMonth, endYear, endMonth
	}

	endYear = reference.Year()
	endMonth = int(reference.Month())
	startYear = endYear
	startMonth = endMonth - 11
	if startMonth <= 0 {
		startMonth += 12
		startYear--
	}
	return startYear, startMonth, endYear, endMonth
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
