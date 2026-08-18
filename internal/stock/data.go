package stock

type ListOutput struct {
	Stocks []StockOutput `json:"stocks"`
}

type StockOutput struct {
	ID      string  `json:"id"`
	Ticker  string  `json:"ticker"`
	Name    string  `json:"name"`
	Sector  string  `json:"sector"`
	Rank    int8    `json:"rank"`
	Website *string `json:"website"`
}

type SeedInput struct {
	Force bool `json:"force"`
}

type SeedOutput struct {
	Inserted int `json:"inserted"`
	Updated  int `json:"updated"`
	Skipped  int `json:"skipped"`
}
