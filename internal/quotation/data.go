package quotation

type QuoteOutput struct {
	Ticker string `json:"ticker"`
	Price  int64  `json:"price"`
}
