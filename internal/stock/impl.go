package stock

import (
	"context"
	"errors"
)

type stockService struct {
	repo Repository
}

func NewStockService(repo Repository) Service {
	return &stockService{repo: repo}
}

func (s *stockService) ListStocks(ctx context.Context) ([]Stock, error) {
	stocks, err := s.repo.List(ctx)
	if err != nil {
		return nil, ErrFailedToList
	}
	return stocks, nil
}

func (s *stockService) GetStockByTicker(ctx context.Context, ticker string) (*Stock, error) {
	stock, err := s.repo.FindByTicker(ctx, ticker)
	if err != nil {
		if errors.Is(err, ErrStockNotFound) {
			return nil, ErrStockNotFound
		}
		return nil, ErrFailedToFind
	}
	return stock, nil
}

func (s *stockService) SeedStocks(ctx context.Context, force bool) (int, int, int, error) {
	stocks := getSeedData()
	return s.repo.Seed(ctx, stocks, force)
}

func getSeedData() []Stock {
	return []Stock{
		{Ticker: "PETR4", Name: "Petrobras", Sector: "Energia", Rank: 10, Website: strPtr("https://www.petrobras.com.br")},
		{Ticker: "VALE3", Name: "Vale", Sector: "Mineração", Rank: 10, Website: strPtr("https://www.vale.com")},
		{Ticker: "ITUB4", Name: "Itaú Unibanco", Sector: "Finanças", Rank: 9, Website: strPtr("https://www.itau.com.br")},
		{Ticker: "BBDC4", Name: "Bradesco", Sector: "Finanças", Rank: 8, Website: strPtr("https://www.bradesco.com.br")},
		{Ticker: "B3SA3", Name: "B3", Sector: "Finanças", Rank: 9, Website: strPtr("https://www.b3.com.br")},
		{Ticker: "ABEV3", Name: "Ambev", Sector: "Consumo", Rank: 7, Website: strPtr("https://www.ambev.com.br")},
		{Ticker: "MGLU3", Name: "Magalu", Sector: "Varejo", Rank: 6, Website: strPtr("https://www.magazineluiza.com.br")},
		{Ticker: "WEGE3", Name: "WEG", Sector: "Industrial", Rank: 9, Website: strPtr("https://www.weg.com")},
		{Ticker: "RENT3", Name: "Localiza", Sector: "Aluguel", Rank: 8, Website: strPtr("https://www.localiza.com")},
		{Ticker: "PRIO3", Name: "Petrobras Rio", Sector: "Energia", Rank: 7, Website: strPtr("https://www.petrobras.com.br")},
		{Ticker: "EQTL3", Name: "Equatorial", Sector: "Energia", Rank: 7, Website: strPtr("https://www.equatorialenergia.com.br")},
		{Ticker: "SUZB3", Name: "Suzano", Sector: "Papel", Rank: 8, Website: strPtr("https://www.suzano.com.br")},
		{Ticker: "KLBN4", Name: "Klabin", Sector: "Papel", Rank: 7, Website: strPtr("https://www.klabin.com.br")},
		{Ticker: "BBAS3", Name: "Banco do Brasil", Sector: "Finanças", Rank: 9, Website: strPtr("https://www.bb.com.br")},
		{Ticker: "SANB3", Name: "Santander Brasil", Sector: "Finanças", Rank: 8, Website: strPtr("https://www.santander.com.br")},
		{Ticker: "JBSS3", Name: "JBS", Sector: "Alimentos", Rank: 7, Website: strPtr("https://www.jbs.com.br")},
		{Ticker: "RDOR3", Name: "RaiaDrogasil", Sector: "Saúde", Rank: 8, Website: strPtr("https://www.rd.com.br")},
		{Ticker: "HAPV3", Name: "Hapvida", Sector: "Saúde", Rank: 6, Website: strPtr("https://www.hapvida.com.br")},
		{Ticker: "NTCO3", Name: "NTCo", Sector: "Construção", Rank: 6, Website: strPtr("https://www.nteconstrutora.com.br")},
		{Ticker: "CCRO3", Name: "CCR", Sector: "Construção", Rank: 6, Website: strPtr("https://www.ccr.com.br")},
	}
}

func strPtr(s string) *string {
	return &s
}
