package stock

import (
	"context"
	"errors"
	"github.com/google/uuid"
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
		{ID: uuid.New().String(), Ticker: "PETR4", Name: "Petrobras", Sector: "Energia", Rank: 10, Website: strPtr("https://www.petrobras.com.br")},
		{ID: uuid.New().String(), Ticker: "VALE3", Name: "Vale", Sector: "Mineração", Rank: 10, Website: strPtr("https://www.vale.com")},
		{ID: uuid.New().String(), Ticker: "ITUB4", Name: "Itaú Unibanco", Sector: "Finanças", Rank: 9, Website: strPtr("https://www.itau.com.br")},
		{ID: uuid.New().String(), Ticker: "BBDC4", Name: "Bradesco", Sector: "Finanças", Rank: 8, Website: strPtr("https://www.bradesco.com.br")},
		{ID: uuid.New().String(), Ticker: "B3SA3", Name: "B3", Sector: "Finanças", Rank: 9, Website: strPtr("https://www.b3.com.br")},
		{ID: uuid.New().String(), Ticker: "ABEV3", Name: "Ambev", Sector: "Consumo", Rank: 7, Website: strPtr("https://www.ambev.com.br")},
		{ID: uuid.New().String(), Ticker: "MGLU3", Name: "Magalu", Sector: "Varejo", Rank: 6, Website: strPtr("https://www.magazineluiza.com.br")},
		{ID: uuid.New().String(), Ticker: "WEGE3", Name: "WEG", Sector: "Industrial", Rank: 9, Website: strPtr("https://www.weg.com")},
		{ID: uuid.New().String(), Ticker: "RENT3", Name: "Localiza", Sector: "Aluguel", Rank: 8, Website: strPtr("https://www.localiza.com")},
		{ID: uuid.New().String(), Ticker: "PRIO3", Name: "Petrobras Rio", Sector: "Energia", Rank: 7, Website: strPtr("https://www.petrobras.com.br")},
		{ID: uuid.New().String(), Ticker: "EQTL3", Name: "Equatorial", Sector: "Energia", Rank: 7, Website: strPtr("https://www.equatorialenergia.com.br")},
		{ID: uuid.New().String(), Ticker: "SUZB3", Name: "Suzano", Sector: "Papel", Rank: 8, Website: strPtr("https://www.suzano.com.br")},
		{ID: uuid.New().String(), Ticker: "KLBN4", Name: "Klabin", Sector: "Papel", Rank: 7, Website: strPtr("https://www.klabin.com.br")},
		{ID: uuid.New().String(), Ticker: "BBAS3", Name: "Banco do Brasil", Sector: "Finanças", Rank: 9, Website: strPtr("https://www.bb.com.br")},
		{ID: uuid.New().String(), Ticker: "SANB3", Name: "Santander Brasil", Sector: "Finanças", Rank: 8, Website: strPtr("https://www.santander.com.br")},
		{ID: uuid.New().String(), Ticker: "JBSS3", Name: "JBS", Sector: "Alimentos", Rank: 7, Website: strPtr("https://www.jbs.com.br")},
		{ID: uuid.New().String(), Ticker: "RDOR3", Name: "RaiaDrogasil", Sector: "Saúde", Rank: 8, Website: strPtr("https://www.rd.com.br")},
		{ID: uuid.New().String(), Ticker: "HAPV3", Name: "Hapvida", Sector: "Saúde", Rank: 6, Website: strPtr("https://www.hapvida.com.br")},
		{ID: uuid.New().String(), Ticker: "NTCO3", Name: "NTCo", Sector: "Construção", Rank: 6, Website: strPtr("https://www.nteconstrutora.com.br")},
		{ID: uuid.New().String(), Ticker: "CCRO3", Name: "CCR", Sector: "Construção", Rank: 6, Website: strPtr("https://www.ccr.com.br")},
	}
}

func strPtr(s string) *string {
	return &s
}
