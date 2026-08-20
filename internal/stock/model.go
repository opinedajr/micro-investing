package stock

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Stock struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Ticker    string    `json:"ticker" gorm:"uniqueIndex:idx_stocks_ticker" validate:"required,ticker"`
	Name      string    `json:"name" validate:"required,min=2,max=100"`
	Sector    string    `json:"sector" validate:"required,min=2,max=50"`
	Rank      int8      `json:"rank" validate:"required,gte=0,lte=10"`
	Website   *string   `json:"website"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *Stock) BeforeCreate(tx *gorm.DB) error {
	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	return nil
}

type CurrentPrice struct {
	StockID   string    `json:"stock_id" gorm:"primaryKey"`
	Price     int64     `json:"price"`
	UpdatedAt time.Time `json:"updated_at"`
}
