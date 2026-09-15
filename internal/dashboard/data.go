package dashboard

type SummaryOutput struct {
	CurrentPatrimony int64 `json:"current_patrimony"`
	YearlyDividends  int64 `json:"yearly_dividends"`
	StocksInvested   int64 `json:"stocks_invested"`
}
