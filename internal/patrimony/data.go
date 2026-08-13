package patrimony

import "time"

type CreatePatrimonyInput struct {
	WalletID string    `json:"-"`
	Year     int       `json:"year" validate:"required"`
	Month    int       `json:"month" validate:"required"`
	Type     AssetType `json:"type" validate:"required"`
	Amount   int64     `json:"amount" validate:"required"`
}

type UpdatePatrimonyInput struct {
	WalletID string    `json:"-"`
	Year     int       `json:"year" validate:"required"`
	Month    int       `json:"month" validate:"required"`
	Type     AssetType `json:"type" validate:"required"`
	Amount   int64     `json:"amount" validate:"required"`
}

type PatrimonyOutput struct {
	ID        string    `json:"id"`
	WalletID  string    `json:"wallet_id"`
	Year      int       `json:"year"`
	Month     int       `json:"month"`
	Type      AssetType `json:"type"`
	Amount    int64     `json:"amount"`
	CreatedAt string    `json:"created_at"`
	UpdatedAt string    `json:"updated_at"`
}

type PatrimonyFilter struct {
	WalletID string
	Type     AssetType
	Year     int
	Month    int
}

type CreateAssetInput struct {
	WalletID    string    `json:"-"`
	Type        AssetType `json:"type" validate:"required"`
	Date        string    `json:"date" validate:"required"`
	Description string    `json:"description" validate:"required"`
	Amount      int64     `json:"amount" validate:"required"`
}

type UpdateAssetInput struct {
	WalletID    string    `json:"-"`
	Type        AssetType `json:"type" validate:"required"`
	Date        string    `json:"date" validate:"required"`
	Description string    `json:"description" validate:"required"`
	Amount      int64     `json:"amount" validate:"required"`
}

type AssetOutput struct {
	ID          string    `json:"id"`
	WalletID    string    `json:"wallet_id"`
	Type        AssetType `json:"type"`
	Date        string    `json:"date"`
	Description string    `json:"description"`
	Amount      int64     `json:"amount"`
	CreatedAt   string    `json:"created_at"`
	UpdatedAt   string    `json:"updated_at"`
}

type AssetFilter struct {
	WalletID  string
	Type      AssetType
	StartDate *time.Time
	EndDate   *time.Time
}
