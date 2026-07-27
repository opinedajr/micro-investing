package wallet

type CreateWalletInput struct {
	UserID      string
	Name        string
	Description *string
}

type WalletOutput struct {
	ID          string
	Name        string
	Description *string
	CreatedAt    string
	UpdatedAt    string
}
