package stock

var BlueChips = func() []Stock {
	return []Stock{
		{Ticker: "PETR4", Name: "Petrobras PN", Sector: "Petróleo, Gás e Biocombustíveis", Rank: 10, Website: ptr("https://petrobras.com.br")},
		{Ticker: "VALE3", Name: "Vale ON", Sector: "Mineração e Siderurgia", Rank: 10, Website: ptr("https://vale.com")},
		{Ticker: "ITUB4", Name: "Itaú Unibanco PN", Sector: "Bancos", Rank: 10, Website: ptr("https://itau.com.br")},
		{Ticker: "BBDC4", Name: "Bradesco PN", Sector: "Bancos", Rank: 9, Website: ptr("https://banco.bradesco")},
		{Ticker: "B3SA3", Name: "B3 ON", Sector: "Bolsa de Valores e Serviços Financeiros", Rank: 9, Website: ptr("https://b3.com.br")},
		{Ticker: "ABEV3", Name: "Ambev ON", Sector: "Bebidas", Rank: 9, Website: ptr("https://ambev.com.br")},
		{Ticker: "MGLU3", Name: "Magazine Luiza ON", Sector: "Varejo", Rank: 6, Website: ptr("https://magazineluiza.com.br")},
		{Ticker: "BBSE3", Name: "BB Seguridade ON", Sector: "Seguradoras", Rank: 9, Website: ptr("https://bbseguridade.com.br")},
		{Ticker: "SANB11", Name: "Santander Brasil Unt", Sector: "Bancos", Rank: 9, Website: ptr("https://www.santander.com.br")},
		{Ticker: "ITSA4", Name: "Itaúsa PN", Sector: "Holding Diversificada", Rank: 9, Website: ptr("https://itausa.com.br")},
		{Ticker: "WEGE3", Name: "WEG ON", Sector: "Máquinas e Equipamentos", Rank: 10, Website: ptr("https://weg.net")},
		{Ticker: "RENT3", Name: "Localiza ON", Sector: "Locação de Veículos", Rank: 9, Website: ptr("https://localiza.com")},
		{Ticker: "SUZB3", Name: "Suzano ON", Sector: "Papel e Celulose", Rank: 9, Website: ptr("https://suzano.com.br")},
		{Ticker: "EMBR3", Name: "Embraer ON", Sector: "Material de Transporte", Rank: 8, Website: ptr("https://embraer.com")},
		{Ticker: "JBSS3", Name: "JBS ON", Sector: "Alimentos Processados", Rank: 8, Website: ptr("https://jbs.com.br")},
		{Ticker: "VIVT3", Name: "Telefônica Brasil ON", Sector: "Telecomunicações", Rank: 9, Website: ptr("https://vivo.com.br")},
		{Ticker: "CMIG4", Name: "CEMIG PN", Sector: "Energia Elétrica", Rank: 8, Website: ptr("https://cemig.com.br")},
		{Ticker: "ELET3", Name: "Eletrobras ON", Sector: "Energia Elétrica", Rank: 8, Website: ptr("https://eletrobras.com")},
		{Ticker: "EQTL3", Name: "Equatorial Energia ON", Sector: "Energia Elétrica", Rank: 8, Website: ptr("https://equatorialenergia.com.br")},
		{Ticker: "GGBR4", Name: "Gerdau PN", Sector: "Mineração e Siderurgia", Rank: 8, Website: ptr("https://gerdau.com")},
		{Ticker: "PRIO3", Name: "Petro Rio ON", Sector: "Petróleo, Gás e Biocombustíveis", Rank: 7, Website: ptr("https://petrorio.com.br")},
		{Ticker: "RDOR3", Name: "Rede D'Or ON", Sector: "Saúde", Rank: 8, Website: ptr("https://rededor.com.br")},
		{Ticker: "TOTS3", Name: "Totvs ON", Sector: "Software e Serviços de TI", Rank: 8, Website: ptr("https://totvs.com.br")},
		{Ticker: "LREN3", Name: "Lojas Renner ON", Sector: "Varejo", Rank: 7, Website: ptr("https://lojasrenner.com.br")},
		{Ticker: "HAPV3", Name: "Hapvida ON", Sector: "Saúde", Rank: 5, Website: ptr("https://hapvida.com.br")},
		{Ticker: "ENBR3", Name: "Energias do Brasil ON", Sector: "Energia Elétrica", Rank: 7, Website: ptr("https://energiasdobrasil.com.br")},
		{Ticker: "CSAN3", Name: "Cosan ON", Sector: "Petróleo, Gás e Biocombustíveis", Rank: 7, Website: ptr("https://cosan.com.br")},
		{Ticker: "TIMS3", Name: "TIM ON", Sector: "Telecomunicações", Rank: 8, Website: ptr("https://tim.com.br")},
		{Ticker: "BBDC3", Name: "Bradesco ON", Sector: "Bancos", Rank: 8, Website: ptr("https://banco.bradesco")},
	}
}

func ptr(s string) *string {
	return &s
}
