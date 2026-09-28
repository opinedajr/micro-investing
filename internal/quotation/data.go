package quotation

type QuoteOutput struct {
	Ticker string `json:"ticker"`
	Price  int64  `json:"price"`
}

type SyncInput struct {
	Tickers []string `json:"tickers"`
}

type SyncOutput struct {
	Updated int      `json:"updated"`
	Skipped []string `json:"skipped"`
	Failed  []string `json:"failed"`
}
