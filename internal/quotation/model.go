package quotation

import "time"

type CurrentPrice struct {
	StockID   string    `json:"stock_id" gorm:"primaryKey"`
	Price     int64     `json:"price"`
	UpdatedAt time.Time `json:"updated_at"`
}
