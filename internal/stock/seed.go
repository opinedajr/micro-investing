package stock

var BlueChips = func() []Stock {
	return []Stock{
		{Ticker: "BBDC3", Name: "Bradesco ON", Sector: "Bancos", Rank: 5, Website: ptr("https://statusinvest.com.br/acoes/bbdc3")},
		{Ticker: "PETR4", Name: "Petrobras PN", Sector: "Commodities Cíclicas", Rank: 4, Website: ptr("https://statusinvest.com.br/acoes/petr4")},
		{Ticker: "VALE3", Name: "Vale ON", Sector: "Commodities Cíclicas", Rank: 5, Website: ptr("https://statusinvest.com.br/acoes/vale3")},
		{Ticker: "ITUB3", Name: "Itaú Unibanco ON", Sector: "Bancos", Rank: 5, Website: ptr("https://statusinvest.com.br/acoes/itub3")},
		{Ticker: "B3SA3", Name: "B3 ON", Sector: "Financeiro", Rank: 4, Website: ptr("https://statusinvest.com.br/acoes/b3sa3")},
		{Ticker: "ABEV3", Name: "Ambev ON", Sector: "Consumo", Rank: 4, Website: ptr("https://statusinvest.com.br/acoes/abev3")},
		{Ticker: "BBSE3", Name: "BB Seguridade ON", Sector: "Seguros", Rank: 4, Website: ptr("https://statusinvest.com.br/acoes/bbse3")},
		{Ticker: "SANB3", Name: "Santander Brasil ON", Sector: "Bancos", Rank: 5, Website: ptr("https://statusinvest.com.br/acoes/sanb3")},
		{Ticker: "ITSA4", Name: "Itaúsa PN", Sector: "Financeiro", Rank: 5, Website: ptr("https://statusinvest.com.br/acoes/itsa4")},
		{Ticker: "WEGE3", Name: "WEG ON", Sector: "Bens Industriais", Rank: 4, Website: ptr("https://statusinvest.com.br/acoes/wege3")},
		{Ticker: "SUZB3", Name: "Suzano ON", Sector: "Commodities Cíclicas", Rank: 3, Website: ptr("https://statusinvest.com.br/acoes/suzb3")},
		{Ticker: "CMIG4", Name: "CEMIG PN", Sector: "Energia Elétrica", Rank: 4, Website: ptr("https://statusinvest.com.br/acoes/cmig4")},
		{Ticker: "EGIE3", Name: "Engie Brasil ON", Sector: "Energia Elétrica", Rank: 4, Website: ptr("https://statusinvest.com.br/acoes/egie3")},
		{Ticker: "TAEE4", Name: "Taesa PN", Sector: "Energia Elétrica", Rank: 4, Website: ptr("https://statusinvest.com.br/acoes/taee4")},
		{Ticker: "GGBR4", Name: "Gerdau PN", Sector: "Commodities Cíclicas", Rank: 3, Website: ptr("https://statusinvest.com.br/acoes/ggbr4")},
		{Ticker: "BBAS3", Name: "Banco do Brasil ON", Sector: "Bancos", Rank: 5, Website: ptr("https://statusinvest.com.br/acoes/bbas3")},
		{Ticker: "CXSE3", Name: "Caixa Seguridade ON", Sector: "Seguros", Rank: 4, Website: ptr("https://statusinvest.com.br/acoes/cxse3")},
		{Ticker: "BRSR6", Name: "Banco Banrisul PN", Sector: "Bancos", Rank: 3, Website: ptr("https://statusinvest.com.br/acoes/brsr6")},
		{Ticker: "KLBN11", Name: "Klabin ON", Sector: "Commodities Cíclicas", Rank: 3, Website: ptr("https://statusinvest.com.br/acoes/klbn11")},
		{Ticker: "PSSA3", Name: "Porto Seguro ON", Sector: "Seguros", Rank: 5, Website: ptr("https://statusinvest.com.br/acoes/pssa3")},
	}
}

func ptr(s string) *string {
	return &s
}
