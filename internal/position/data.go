package position

type CreatePositionInput struct {
	WalletID     string `json:"-"`
	StockID      string `json:"stock_id" validate:"required"`
	Quantity     int64  `json:"quantity" validate:"required,min=1"`
	AveragePrice int64  `json:"average_price" validate:"required,min=1"`
}

type UpdatePositionInput struct {
	WalletID     string `json:"-"`
	PositionID   string `json:"-"`
	Quantity     int64  `json:"quantity" validate:"required,min=1"`
	AveragePrice int64  `json:"average_price" validate:"required,min=1"`
}

type PositionOutput struct {
	ID               string  `json:"id"`
	WalletID         string  `json:"wallet_id"`
	StockID          string  `json:"stock_id"`
	Quantity         int64   `json:"quantity"`
	AveragePrice     int64   `json:"average_price"`
	CurrentPrice     int64   `json:"current_price"`
	Invested         int64   `json:"invested"`
	Balance          int64   `json:"balance"`
	VariationValue   int64   `json:"variation_value"`
	VariationPercent float64 `json:"variation_percent"`
	PortfolioPercent float64 `json:"portfolio_percent"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

type PositionFilter struct {
	WalletID string
	Ticker   string
	Sort     string
}
