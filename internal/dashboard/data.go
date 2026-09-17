package dashboard

type SummaryOutput struct {
	CurrentPatrimony int64 `json:"current_patrimony"`
	YearlyDividends  int64 `json:"yearly_dividends"`
	StocksInvested   int64 `json:"stocks_invested"`
}

type AllocationItem struct {
	Type       string  `json:"type"`
	Amount     int64   `json:"amount"`
	Percentage float64 `json:"percentage"`
}

type AllocationOutput struct {
	Items []AllocationItem `json:"items"`
	Total int64            `json:"total"`
}

type RiskItem struct {
	Rank       int8    `json:"rank"`
	Amount     int64   `json:"amount"`
	Percentage float64 `json:"percentage"`
}

type RiskOutput struct {
	Items []RiskItem `json:"items"`
	Total int64      `json:"total"`
}

type EvolutionMonthOutput struct {
	Year   int   `json:"year"`
	Month  int   `json:"month"`
	Amount int64 `json:"amount"`
}

type EvolutionCategoryOutput struct {
	FixedIncome      []EvolutionMonthOutput `json:"fixed_income"`
	Stocks           []EvolutionMonthOutput `json:"stocks"`
	EmergencyReserve []EvolutionMonthOutput `json:"emergency_reserve"`
}

type EvolutionOutput struct {
	Total      []EvolutionMonthOutput  `json:"total"`
	ByCategory EvolutionCategoryOutput `json:"by_category"`
}

type EvolutionInput struct {
	Year    int
	Quarter int
}
