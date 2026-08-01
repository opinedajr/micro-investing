package wallet

type CreateWalletInput struct {
	UserID      string  `json:"-"`
	Name        string  `json:"name" validate:"required,min=3,max=50"`
	Description *string `json:"description" validate:"omitempty,max=255"`
}

type UpdateWalletInput struct {
	Name        string  `json:"name" validate:"required,min=3,max=50"`
	Description *string `json:"description" validate:"omitempty,max=255"`
}

type WalletOutput struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}
