package position

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Position struct {
	ID               string    `json:"id" gorm:"primaryKey"`
	WalletID         string    `json:"wallet_id" gorm:"uniqueIndex:idx_positions_wallet_stock"`
	StockID          string    `json:"stock_id" gorm:"uniqueIndex:idx_positions_wallet_stock"`
	Quantity         int64     `json:"quantity"`
	AveragePrice     int64     `json:"average_price"`
	CurrentPrice     int64     `json:"current_price"`
	Invested         int64     `json:"invested"`
	Balance          int64     `json:"balance"`
	VariationValue   int64     `json:"variation_value"`
	VariationPercent float64   `json:"variation_percent"`
	PortfolioPercent float64   `json:"portfolio_percent"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func (p *Position) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}
