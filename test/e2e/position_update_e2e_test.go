//go:build integration

package e2e

import (
	"net/http"

	"github.com/opinedajr/micro-investing/internal/position"
)

func (s *E2ESuite) positionUpdateFixture() (walletID, petr4PositionID, vale3PositionID, petr4StockID, vale3StockID string) {
	walletID = s.positionWalletID("Carteira Update")
	petr4StockID = s.positionStockID("PETR4")
	vale3StockID = s.positionStockID("VALE3")

	s.positionInsertPrice(petr4StockID, 7500)
	s.positionInsertPrice(vale3StockID, 12000)

	petr4PositionID = s.positionCreate(walletID, petr4StockID, 100, 5000)
	vale3PositionID = s.positionCreate(walletID, vale3StockID, 100, 10000)

	return walletID, petr4PositionID, vale3PositionID, petr4StockID, vale3StockID
}

func (s *E2ESuite) TestPosition_Update_Success() {
	walletID, petr4PositionID, _, petr4StockID, vale3StockID := s.positionUpdateFixture()

	resp := s.expect.PUT("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", petr4PositionID).
		WithJSON(position.UpdatePositionInput{
			Quantity:     200,
			AveragePrice: 6000,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	resp.Value("id").String().IsEqual(petr4PositionID)
	resp.Value("wallet_id").String().IsEqual(walletID)
	resp.Value("stock_id").String().IsEqual(petr4StockID)
	resp.Value("quantity").Number().IsEqual(200)
	resp.Value("average_price").Number().IsEqual(6000)
	resp.Value("current_price").Number().IsEqual(7500)
	resp.Value("invested").Number().IsEqual(1200000)
	resp.Value("balance").Number().IsEqual(1500000)
	resp.Value("variation_value").Number().IsEqual(300000)
	resp.Value("variation_percent").Number().IsEqual(25)
	resp.Value("portfolio_percent").Number().InRange(54.54, 54.55)

	arr := s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	var sum float64
	percents := make(map[string]float64)
	for _, item := range arr.Iter() {
		stockID := item.Object().Value("stock_id").String().Raw()
		percent := item.Object().Value("portfolio_percent").Number().Raw()
		percents[stockID] = percent
		sum += percent
	}

	s.Require().Len(percents, 2)
	s.InDelta(54.5454, percents[petr4StockID], 0.001)
	s.InDelta(45.4545, percents[vale3StockID], 0.001)
	s.InDelta(100.0, sum, 0.001)
}

func (s *E2ESuite) TestPosition_Update_InvalidPayload() {
	walletID, petr4PositionID, _, _, _ := s.positionUpdateFixture()

	s.expect.PUT("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", petr4PositionID).
		WithJSON(map[string]interface{}{
			"quantity":      0,
			"average_price": 1000,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().
		Value("code").String().IsEqual("VALIDATION_ERROR")

	s.expect.PUT("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", petr4PositionID).
		WithJSON(map[string]interface{}{
			"quantity":      10,
			"average_price": 0,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().
		Value("code").String().IsEqual("VALIDATION_ERROR")

	s.expect.PUT("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", petr4PositionID).
		WithJSON(map[string]interface{}{}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().
		Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestPosition_Update_PositionNotFound() {
	walletID, _, _, _, _ := s.positionUpdateFixture()

	s.expect.PUT("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", "123e4567-e89b-12d3-a456-426614174000").
		WithJSON(position.UpdatePositionInput{
			Quantity:     10,
			AveragePrice: 1000,
		}).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().
		Value("code").String().IsEqual("POSITION_NOT_FOUND")
}

func (s *E2ESuite) TestPosition_Update_CrossWallet() {
	walletA, petr4PositionID, _, _, _ := s.positionUpdateFixture()
	walletB := s.positionWalletID("Carteira Update B")

	s.expect.PUT("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletB).
		WithPath("positionId", petr4PositionID).
		WithJSON(position.UpdatePositionInput{
			Quantity:     10,
			AveragePrice: 1000,
		}).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().
		Value("code").String().IsEqual("POSITION_NOT_FOUND")

	s.Require().NotEqual(walletA, walletB)
}

func (s *E2ESuite) TestPosition_Update_WalletNotFound() {
	s.expect.PUT("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		WithPath("positionId", "123e4567-e89b-12d3-a456-426614174000").
		WithJSON(position.UpdatePositionInput{
			Quantity:     10,
			AveragePrice: 1000,
		}).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().
		Value("code").String().IsEqual("WALLET_NOT_FOUND")
}

func (s *E2ESuite) TestPosition_Update_StockIDImmutable() {
	walletID, petr4PositionID, _, petr4StockID, vale3StockID := s.positionUpdateFixture()

	resp := s.expect.PUT("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", petr4PositionID).
		WithJSON(map[string]interface{}{
			"quantity":      300,
			"average_price": 7000,
			"stock_id":      vale3StockID,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	resp.Value("stock_id").String().IsEqual(petr4StockID)
	resp.Value("quantity").Number().IsEqual(300)
	resp.Value("average_price").Number().IsEqual(7000)
}

func (s *E2ESuite) TestPosition_Update_Regression() {
	walletID, petr4PositionID, _, _, _ := s.positionUpdateFixture()

	s.expect.PUT("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", petr4PositionID).
		WithJSON(position.UpdatePositionInput{
			Quantity:     50,
			AveragePrice: 4000,
		}).
		Expect().
		Status(http.StatusOK)

	found := s.expect.GET("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", petr4PositionID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	found.Value("quantity").Number().IsEqual(50)
	found.Value("average_price").Number().IsEqual(4000)
	found.Value("invested").Number().IsEqual(200000)

	list := s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	list.Length().IsEqual(2)

	newPositionID := s.positionCreate(walletID, s.positionStockID("ITUB4"), 5, 2000)
	s.Require().NotEmpty(newPositionID)

	s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array().
		Length().IsEqual(3)
}
