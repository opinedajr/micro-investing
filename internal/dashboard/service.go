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
	Risk(ctx context.Context, walletID string) (*RiskOutput, error)
	Evolution(ctx context.Context, walletID string, input EvolutionInput) (*EvolutionOutput, error)
	Dividends(ctx context.Context, walletID string) (*DividendsOutput, error)
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
	startYear, startMonth, endYear, endMonth, err := resolveEvolutionPeriod(time.Now(), input)
	if err != nil {
		return nil, err
	}

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
		if len(typeAmounts) > 0 {
			for _, item := range typeAmounts {
				monthTotal += item.Amount
			}
			previousTotal = monthTotal
		} else {
			monthTotal = previousTotal
		}

		monthFixedIncome, fixedIncomeFound := categoryAmount(typeAmounts, patrimony.TypeFixedIncome)
		monthStocks, stocksFound := categoryAmount(typeAmounts, patrimony.TypeStocks)
		monthEmergencyReserve, emergencyReserveFound := categoryAmount(typeAmounts, patrimony.TypeEmergencyReserve)

		previousFixedIncome = carryForwardValue(previousFixedIncome, monthFixedIncome, fixedIncomeFound)
		previousStocks = carryForwardValue(previousStocks, monthStocks, stocksFound)
		previousEmergencyReserve = carryForwardValue(previousEmergencyReserve, monthEmergencyReserve, emergencyReserveFound)

		totalSeries = append(totalSeries, EvolutionMonthOutput{Year: year, Month: month, Amount: monthTotal})
		fixedIncomeSeries = append(fixedIncomeSeries, EvolutionMonthOutput{Year: year, Month: month, Amount: previousFixedIncome})
		stocksSeries = append(stocksSeries, EvolutionMonthOutput{Year: year, Month: month, Amount: previousStocks})
		emergencyReserveSeries = append(emergencyReserveSeries, EvolutionMonthOutput{Year: year, Month: month, Amount: previousEmergencyReserve})

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

func categoryAmount(typeAmounts []patrimony.TypeAmount, assetType patrimony.AssetType) (int64, bool) {
	for _, item := range typeAmounts {
		if item.Type == assetType {
			return item.Amount, true
		}
	}
	return 0, false
}

func carryForwardValue(previous int64, current int64, found bool) int64 {
	if found {
		return current
	}
	return previous
}

func resolveEvolutionPeriod(reference time.Time, input EvolutionInput) (startYear int, startMonth int, endYear int, endMonth int, err error) {
	if input.Year > 0 {
		if input.Quarter != 0 {
			if input.Quarter < 1 || input.Quarter > 4 {
				return 0, 0, 0, 0, ErrInvalidEvolutionPeriod
			}
			startMonth = (input.Quarter-1)*3 + 1
			endMonth = input.Quarter * 3
			startYear = input.Year
			endYear = input.Year
			return startYear, startMonth, endYear, endMonth, nil
		}
		startYear = input.Year
		endYear = input.Year
		startMonth = 1
		endMonth = 12
		return startYear, startMonth, endYear, endMonth, nil
	}

	if input.Quarter != 0 {
		return 0, 0, 0, 0, ErrInvalidEvolutionPeriod
	}

	endYear = reference.Year()
	endMonth = int(reference.Month())
	startYear = endYear
	startMonth = endMonth - 11
	if startMonth <= 0 {
		startMonth += 12
		startYear--
	}
	return startYear, startMonth, endYear, endMonth, nil
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

func (s *dashboardService) Risk(ctx context.Context, walletID string) (*RiskOutput, error) {
	positions, err := s.positionRepository.FindByFilter(ctx, position.PositionFilter{WalletID: walletID})
	if err != nil {
		return nil, err
	}

	stockIDs := make([]string, 0, len(positions))
	for _, p := range positions {
		stockIDs = append(stockIDs, p.StockID)
	}

	stocks, err := s.stockRepository.FindByIDs(ctx, stockIDs)
	if err != nil {
		return nil, err
	}

	rankByStockID := make(map[string]int8)
	for _, s := range stocks {
		rankByStockID[s.ID] = s.Rank
	}

	amountByRank := make(map[int8]int64)
	var total int64
	for _, p := range positions {
		rank, ok := rankByStockID[p.StockID]
		if !ok {
			continue
		}
		amountByRank[rank] += p.Invested
		total += p.Invested
	}

	items := make([]RiskItem, 0, len(amountByRank))
	for rank, amount := range amountByRank {
		if amount <= 0 {
			continue
		}
		percentage := 0.0
		if total > 0 {
			percentage = float64(amount) / float64(total) * 100
		}
		items = append(items, RiskItem{
			Rank:       rank,
			Amount:     amount,
			Percentage: percentage,
		})
	}

	return &RiskOutput{
		Items: items,
		Total: total,
	}, nil
}

func (s *dashboardService) Dividends(ctx context.Context, walletID string) (*DividendsOutput, error) {
	return &DividendsOutput{Items: []DividendItem{}}, nil
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
