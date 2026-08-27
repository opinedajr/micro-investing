//go:build integration

package e2e

import (
	"net/http"
	"time"

	"github.com/opinedajr/micro-investing/internal/position"
)

func (s *E2ESuite) TestPosition_Create_WithCurrentPrice() {
	walletID := s.positionWalletID("Carteira Posições")
	stockID := s.positionStockID("PETR4")
	s.positionInsertPrice(stockID, 7500)

	payload := position.CreatePositionInput{
		StockID:      stockID,
		Quantity:     100,
		AveragePrice: 5000,
	}

	resp := s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(payload).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	resp.Value("id").String().NotEmpty()
	resp.Value("wallet_id").String().IsEqual(walletID)
	resp.Value("stock_id").String().IsEqual(stockID)
	resp.Value("quantity").Number().IsEqual(100)
	resp.Value("average_price").Number().IsEqual(5000)
	resp.Value("current_price").Number().IsEqual(7500)
	resp.Value("invested").Number().IsEqual(500000)
	resp.Value("balance").Number().IsEqual(750000)
	resp.Value("variation_value").Number().IsEqual(250000)
	resp.Value("variation_percent").Number().IsEqual(50)
	resp.Value("portfolio_percent").Number().IsEqual(100)
	resp.Value("created_at").String().NotEmpty()
	resp.Value("updated_at").String().NotEmpty()
}

func (s *E2ESuite) TestPosition_Create_MissingCurrentPrice() {
	walletID := s.positionWalletID("Carteira Sem Preço")
	stockID := s.positionStockID("PETR4")

	payload := position.CreatePositionInput{
		StockID:      stockID,
		Quantity:     10,
		AveragePrice: 1000,
	}

	resp := s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(payload).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	resp.Value("current_price").Number().IsEqual(0)
	resp.Value("invested").Number().IsEqual(10000)
	resp.Value("balance").Number().IsEqual(0)
	resp.Value("variation_value").Number().IsEqual(-10000)
	resp.Value("variation_percent").Number().IsEqual(-100)
	resp.Value("portfolio_percent").Number().IsEqual(100)
}

func (s *E2ESuite) TestPosition_Create_MultipleRedistributesPortfolioPercent() {
	walletID := s.positionWalletID("Carteira Diversificada")
	petr4ID := s.positionStockID("PETR4")
	vale3ID := s.positionStockID("VALE3")
	s.positionInsertPrice(petr4ID, 7500)
	s.positionInsertPrice(vale3ID, 12000)

	first := position.CreatePositionInput{
		StockID:      petr4ID,
		Quantity:     100,
		AveragePrice: 5000,
	}
	s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(first).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object().
		Value("portfolio_percent").Number().IsEqual(100)

	second := position.CreatePositionInput{
		StockID:      vale3ID,
		Quantity:     100,
		AveragePrice: 10000,
	}
	secondResp := s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(second).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	secondResp.Value("invested").Number().IsEqual(1000000)
	secondResp.Value("balance").Number().IsEqual(1200000)
	secondResp.Value("variation_value").Number().IsEqual(200000)
	secondResp.Value("variation_percent").Number().IsEqual(20)
	secondResp.Value("portfolio_percent").Number().InRange(66.66, 66.67)

	var petr4Percent float64
	err := s.container.DB().
		Table("positions").
		Select("portfolio_percent").
		Where("wallet_id = ? AND stock_id = ?", walletID, petr4ID).
		Scan(&petr4Percent).Error
	s.Require().NoError(err)
	s.InDelta(33.3333, petr4Percent, 0.001)
}

func (s *E2ESuite) TestPosition_Create_WalletNotFound() {
	stockID := s.positionStockID("PETR4")

	payload := position.CreatePositionInput{
		StockID:      stockID,
		Quantity:     10,
		AveragePrice: 1000,
	}

	s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		WithJSON(payload).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}

func (s *E2ESuite) TestPosition_Create_StockNotFound() {
	walletID := s.positionWalletID("Carteira Stock Inexistente")

	payload := position.CreatePositionInput{
		StockID:      "123e4567-e89b-12d3-a456-426614174000",
		Quantity:     10,
		AveragePrice: 1000,
	}

	s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(payload).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("STOCK_NOT_FOUND")
}

func (s *E2ESuite) TestPosition_Create_Duplicate() {
	walletID := s.positionWalletID("Carteira Duplicada")
	stockID := s.positionStockID("PETR4")

	payload := position.CreatePositionInput{
		StockID:      stockID,
		Quantity:     10,
		AveragePrice: 1000,
	}

	s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(payload).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(payload).
		Expect().
		Status(http.StatusConflict).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("POSITION_ALREADY_EXISTS")
}

func (s *E2ESuite) TestPosition_Create_InvalidPayload() {
	walletID := s.positionWalletID("Carteira Validação")
	stockID := s.positionStockID("PETR4")

	s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(map[string]interface{}{
			"stock_id":      stockID,
			"quantity":      0,
			"average_price": 1000,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().
		Value("code").String().IsEqual("VALIDATION_ERROR")

	s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(map[string]interface{}{
			"stock_id":      stockID,
			"quantity":      10,
			"average_price": 0,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().
		Value("code").String().IsEqual("VALIDATION_ERROR")

	s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(map[string]interface{}{
			"stock_id": stockID,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().
		Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) positionWalletID(name string) string {
	resp := s.expect.POST("/api/v1/wallets").
		WithJSON(map[string]interface{}{"name": name}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	return resp.Value("id").String().Raw()
}

func (s *E2ESuite) positionStockID(ticker string) string {
	var id string
	err := s.container.DB().
		Table("stocks").
		Select("id").
		Where("ticker = ?", ticker).
		Scan(&id).Error
	s.Require().NoError(err)
	return id
}

func (s *E2ESuite) positionInsertPrice(stockID string, price int64) {
	err := s.container.DB().
		Exec("INSERT INTO stocks_current_prices (stock_id, price, updated_at) VALUES (?, ?, ?)", stockID, price, time.Now()).Error
	s.Require().NoError(err)
}
