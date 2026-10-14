package dividends

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Dividend struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	WalletID  string    `json:"wallet_id" gorm:"uniqueIndex:idx_dividends_wallet_year"`
	Year      int       `json:"year" gorm:"uniqueIndex:idx_dividends_wallet_year"`
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (d *Dividend) BeforeCreate(tx *gorm.DB) error {
	if d.ID == "" {
		d.ID = uuid.New().String()
	}
	return nil
}
