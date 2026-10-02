//go:build integration

package e2e

import (
	"net/http"

	"github.com/opinedajr/micro-investing/internal/position"
)

type seededStock struct {
	id     string
	ticker string
	name   string
	sector string
	rank   int8
}

func (s *E2ESuite) seededStockByTicker(ticker string) seededStock {
	var row struct {
		ID     string
		Ticker string
		Name   string
		Sector string
		Rank   int8
	}
	err := s.container.DB().
		Table("stocks").
		Select("id, ticker, name, sector, rank").
		Where("ticker = ?", ticker).
		Scan(&row).Error
	s.Require().NoError(err)
	return seededStock{id: row.ID, ticker: row.Ticker, name: row.Name, sector: row.Sector, rank: row.Rank}
}

func (s *E2ESuite) TestPosition_StockSnapshot_AllEndpoints() {
	walletID := s.positionWalletID("Carteira Snapshot")
	petr4 := s.seededStockByTicker("PETR4")
	vale3 := s.seededStockByTicker("VALE3")
	s.positionInsertPrice(petr4.id, 7500)
	s.positionInsertPrice(vale3.id, 12000)

	createResp := s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(position.CreatePositionInput{StockID: petr4.id, Quantity: 100, AveragePrice: 5000}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	stock := createResp.Value("stock").Object()
	stock.Value("id").String().IsEqual(petr4.id)
	stock.Value("ticker").String().IsEqual(petr4.ticker)
	stock.Value("name").String().IsEqual(petr4.name)
	stock.Value("sector").String().IsEqual(petr4.sector)
	stock.Value("rank").Number().IsEqual(float64(petr4.rank))
	createResp.Value("stock_id").String().IsEqual(petr4.id)

	vale3PositionID := s.positionCreate(walletID, vale3.id, 10, 1000)

	listArr := s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	for _, item := range listArr.Iter() {
		data := item.Object()
		stockID := data.Value("stock_id").String().Raw()
		stockObj := data.Value("stock").Object()
		stockObj.Value("id").String().IsEqual(stockID)
		if stockID == petr4.id {
			stockObj.Value("ticker").String().IsEqual("PETR4")
			stockObj.Value("name").String().IsEqual(petr4.name)
			stockObj.Value("sector").String().IsEqual(petr4.sector)
			stockObj.Value("rank").Number().IsEqual(float64(petr4.rank))
		} else {
			stockObj.Value("ticker").String().IsEqual("VALE3")
			stockObj.Value("rank").Number().IsEqual(float64(vale3.rank))
		}
	}

	findResp := s.expect.GET("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", vale3PositionID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	findStock := findResp.Value("stock").Object()
	findStock.Value("id").String().IsEqual(vale3.id)
	findStock.Value("ticker").String().IsEqual("VALE3")
	findStock.Value("name").String().IsEqual(vale3.name)

	updateResp := s.expect.PUT("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", vale3PositionID).
		WithJSON(map[string]interface{}{"quantity": 20, "average_price": 1500}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	updateResp.Value("stock_id").String().IsEqual(vale3.id)
	updateStock := updateResp.Value("stock").Object()
	updateStock.Value("ticker").String().IsEqual("VALE3")
	updateStock.Value("sector").String().IsEqual(vale3.sector)
}

func (s *E2ESuite) TestPosition_StockSnapshot_MissingStockReturnsNull() {
	walletID := s.positionWalletID("Carteira Órfã")
	petr4 := s.seededStockByTicker("PETR4")
	s.positionInsertPrice(petr4.id, 7500)

	positionID := s.positionCreate(walletID, petr4.id, 10, 1000)

	orphanID := "11111111-2222-3333-4444-555555555555"
	err := s.container.DB().
		Exec("UPDATE positions SET stock_id = ? WHERE id = ?", orphanID, positionID).Error
	s.Require().NoError(err)

	findResp := s.expect.GET("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", positionID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	findResp.Value("stock_id").String().IsEqual(orphanID)
	findResp.Value("stock").IsNull()

	listArr := s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	listArr.Length().IsEqual(1)
	listItem := listArr.First().Object()
	listItem.Value("stock_id").String().IsEqual(orphanID)
	listItem.Value("stock").IsNull()

	updateResp := s.expect.PUT("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", positionID).
		WithJSON(map[string]interface{}{"quantity": 30, "average_price": 2000}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	updateResp.Value("stock_id").String().IsEqual(orphanID)
	updateResp.Value("stock").IsNull()
}

func (s *E2ESuite) TestPosition_StockSnapshot_OrphanWithTickerFilterAndSort() {
	walletID := s.positionWalletID("Carteira Órfã Filtro")
	petr4 := s.seededStockByTicker("PETR4")
	vale3 := s.seededStockByTicker("VALE3")
	s.positionInsertPrice(petr4.id, 7500)
	s.positionInsertPrice(vale3.id, 12000)

	validID := s.positionCreate(walletID, petr4.id, 10, 1000)
	orphanPositionID := s.positionCreate(walletID, vale3.id, 5, 2000)

	orphanID := "11111111-2222-3333-4444-555555555555"
	err := s.container.DB().
		Exec("UPDATE positions SET stock_id = ? WHERE id = ?", orphanID, orphanPositionID).Error
	s.Require().NoError(err)

	filtered := s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithQuery("ticker", "petr").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	filtered.Length().IsEqual(1)
	filtered.First().Object().Value("id").String().IsEqual(validID)

	byTicker := s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithQuery("sort", "ticker").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	byTicker.Length().IsEqual(2)
	byTicker.First().Object().Value("stock_id").String().IsEqual(orphanID)
	byTicker.First().Object().Value("stock").IsNull()
	byTicker.Last().Object().Value("stock_id").String().IsEqual(petr4.id)

	byRank := s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithQuery("sort", "rank").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	byRank.Length().IsEqual(2)
	byRank.First().Object().Value("stock_id").String().IsEqual(orphanID)
	byRank.Last().Object().Value("stock_id").String().IsEqual(petr4.id)

	byTickerDesc := s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithQuery("sort", "-ticker").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	byTickerDesc.Length().IsEqual(2)
	byTickerDesc.First().Object().Value("stock_id").String().IsEqual(petr4.id)
	byTickerDesc.Last().Object().Value("stock_id").String().IsEqual(orphanID)
}
