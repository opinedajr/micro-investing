package dividends

type CreateDividendInput struct {
	WalletID string `json:"-"`
	Year     int    `json:"year" validate:"required"`
	Amount   int64  `json:"amount" validate:"required"`
}

type DividendOutput struct {
	ID     string `json:"id"`
	Year   int    `json:"year"`
	Amount int64  `json:"amount"`
}

type DividendsListOutput struct {
	Items []DividendOutput `json:"items"`
}

type DividendFilter struct {
	WalletID string
	Year     int
}
