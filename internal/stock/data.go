package stock

type StockOutput struct {
	ID        string  `json:"id"`
	Ticker    string  `json:"ticker"`
	Name      string  `json:"name"`
	Sector    string  `json:"sector"`
	Rank      int8    `json:"rank"`
	Website   *string `json:"website"`
	CreatedAt string  `json:"created_at"`
	UpdatedAt string  `json:"updated_at"`
}

type SeedInput struct {
	Force bool
}
