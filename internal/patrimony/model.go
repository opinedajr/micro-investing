package patrimony

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AssetType string

const (
	TypeStocks           AssetType = "stocks"
	TypeFIIs             AssetType = "fiis"
	TypeFixedIncome      AssetType = "fixed_income"
	TypeEmergencyReserve AssetType = "emergency_reserve"
	TypeLiquidCash       AssetType = "liquid_cash"
)

func (at AssetType) IsValid() bool {
	switch at {
	case TypeStocks, TypeFIIs, TypeFixedIncome, TypeEmergencyReserve, TypeLiquidCash:
		return true
	}
	return false
}

type Patrimony struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	WalletID  string    `json:"wallet_id" gorm:"uniqueIndex:idx_wallet_year_month_type"`
	Year      int       `json:"year" gorm:"uniqueIndex:idx_wallet_year_month_type"`
	Month     int       `json:"month" gorm:"uniqueIndex:idx_wallet_year_month_type"`
	Type      AssetType `json:"type" gorm:"uniqueIndex:idx_wallet_year_month_type"`
	Amount    int64     `json:"amount"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (p *Patrimony) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.New().String()
	}
	return nil
}

type Asset struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	WalletID    string    `json:"wallet_id"`
	Type        AssetType `json:"type"`
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Amount      int64     `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (a *Asset) BeforeCreate(tx *gorm.DB) error {
	if a.ID == "" {
		a.ID = uuid.New().String()
	}
	return nil
}
