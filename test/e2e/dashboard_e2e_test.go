//go:build integration

package e2e

import (
	"net/http"

	"github.com/opinedajr/micro-investing/internal/position"
)

func (s *E2ESuite) TestDashboard_Summary_Success() {
	walletID := s.createWallet("Carteira Dashboard")

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2025,
			"month":  12,
			"type":   "stocks",
			"amount": 100000,
		}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2025,
			"month":  12,
			"type":   "fiis",
			"amount": 50000,
		}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  3,
			"type":   "stocks",
			"amount": 200000,
		}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  3,
			"type":   "fixed_income",
			"amount": 150000,
		}).
		Expect().
		Status(http.StatusCreated)

	petr4ID := s.positionStockID("PETR4")
	vale3ID := s.positionStockID("VALE3")

	s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(position.CreatePositionInput{
			StockID:      petr4ID,
			Quantity:     100,
			AveragePrice: 5000,
		}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(position.CreatePositionInput{
			StockID:      vale3ID,
			Quantity:     100,
			AveragePrice: 10000,
		}).
		Expect().
		Status(http.StatusCreated)

	summary := s.expect.GET("/api/v1/wallets/{id}/dashboard/summary").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	summary.Value("current_patrimony").Number().IsEqual(350000)
	summary.Value("stocks_invested").Number().IsEqual(1500000)
	summary.Value("yearly_dividends").Number().IsEqual(0)

	summary.ContainsKey("current_patrimony")
	summary.ContainsKey("stocks_invested")
	summary.ContainsKey("yearly_dividends")
	summary.NotContainsKey("currentPatrimony")
	summary.NotContainsKey("stocksInvested")
	summary.NotContainsKey("yearlyDividends")
}

func (s *E2ESuite) TestDashboard_Summary_WalletNotFound() {
	s.expect.GET("/api/v1/wallets/{id}/dashboard/summary").
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}

func (s *E2ESuite) TestDashboard_Summary_WalletWithoutPatrimony() {
	walletID := s.createWallet("Carteira Sem Patrimônio")

	petr4ID := s.positionStockID("PETR4")

	s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(position.CreatePositionInput{
			StockID:      petr4ID,
			Quantity:     100,
			AveragePrice: 5000,
		}).
		Expect().
		Status(http.StatusCreated)

	summary := s.expect.GET("/api/v1/wallets/{id}/dashboard/summary").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	summary.Value("current_patrimony").Number().IsEqual(0)
	summary.Value("stocks_invested").Number().IsEqual(500000)
	summary.Value("yearly_dividends").Number().IsEqual(0)
}

func (s *E2ESuite) TestDashboard_Summary_EmptyWallet() {
	walletID := s.createWallet("Carteira Vazia")

	summary := s.expect.GET("/api/v1/wallets/{id}/dashboard/summary").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	summary.Value("current_patrimony").Number().IsEqual(0)
	summary.Value("stocks_invested").Number().IsEqual(0)
	summary.Value("yearly_dividends").Number().IsEqual(0)
}
