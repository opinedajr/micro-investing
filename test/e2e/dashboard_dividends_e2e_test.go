//go:build integration

package e2e

import (
	"net/http"
	"time"
)

func (s *E2ESuite) createDividend(walletID string, year int, amount int64) {
	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   year,
			"amount": amount,
		}).
		Expect().
		Status(http.StatusCreated)
}

func (s *E2ESuite) TestDashboard_Dividends_Success() {
	walletID := s.createWallet("Carteira Dividends")
	currentYear := time.Now().Year()

	s.createDividend(walletID, currentYear, 350000)
	s.createDividend(walletID, currentYear-2, 100000)
	s.createDividend(walletID, currentYear-1, 250000)

	items := s.expect.GET("/api/v1/wallets/{id}/dashboard/dividends").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object().Value("items").Array()

	items.Length().IsEqual(3)

	first := items.Element(0).Object()
	first.Value("year").Number().IsEqual(float64(currentYear - 2))
	first.Value("amount").Number().IsEqual(100000)
	first.NotContainsKey("id")

	second := items.Element(1).Object()
	second.Value("year").Number().IsEqual(float64(currentYear - 1))
	second.Value("amount").Number().IsEqual(250000)

	third := items.Element(2).Object()
	third.Value("year").Number().IsEqual(float64(currentYear))
	third.Value("amount").Number().IsEqual(350000)
}

func (s *E2ESuite) TestDashboard_Dividends_NoYearWindowLimit() {
	walletID := s.createWallet("Carteira Dividends Sem Janela")
	currentYear := time.Now().Year()

	s.createDividend(walletID, currentYear+1, 999000)
	s.createDividend(walletID, currentYear-5, 50000)
	s.createDividend(walletID, 1900, 1000)

	items := s.expect.GET("/api/v1/wallets/{id}/dashboard/dividends").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object().Value("items").Array()

	items.Length().IsEqual(3)

	first := items.Element(0).Object()
	first.Value("year").Number().IsEqual(1900)
	first.Value("amount").Number().IsEqual(1000)

	second := items.Element(1).Object()
	second.Value("year").Number().IsEqual(float64(currentYear - 5))
	second.Value("amount").Number().IsEqual(50000)

	third := items.Element(2).Object()
	third.Value("year").Number().IsEqual(float64(currentYear + 1))
	third.Value("amount").Number().IsEqual(999000)
	third.NotContainsKey("id")
}

func (s *E2ESuite) TestDashboard_Dividends_Empty() {
	walletID := s.createWallet("Carteira Dividends Vazia")

	s.expect.GET("/api/v1/wallets/{id}/dashboard/dividends").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object().Value("items").Array().Length().IsEqual(0)
}

func (s *E2ESuite) TestDashboard_Dividends_WalletNotFound() {
	s.expect.GET("/api/v1/wallets/{id}/dashboard/dividends").
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}

func (s *E2ESuite) TestDashboard_Summary_YearlyDividends() {
	walletID := s.createWallet("Carteira Summary Dividends")
	currentYear := time.Now().Year()

	s.createDividend(walletID, currentYear, 350000)
	s.createDividend(walletID, currentYear-1, 250000)

	summary := s.expect.GET("/api/v1/wallets/{id}/dashboard/summary").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	summary.Value("yearly_dividends").Number().IsEqual(350000)
	summary.NotContainsKey("dividends_year")
}

func (s *E2ESuite) TestDashboard_Summary_YearlyDividendsFallsBackToPreviousYear() {
	walletID := s.createWallet("Carteira Summary Dividends Anterior")
	currentYear := time.Now().Year()

	s.createDividend(walletID, currentYear-1, 250000)

	summary := s.expect.GET("/api/v1/wallets/{id}/dashboard/summary").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	summary.Value("yearly_dividends").Number().IsEqual(250000)
}

func (s *E2ESuite) TestDashboard_Summary_YearlyDividendsZeroWhenOnlyFutureYear() {
	walletID := s.createWallet("Carteira Summary Dividends Futuro")
	currentYear := time.Now().Year()

	s.createDividend(walletID, currentYear+1, 500000)

	summary := s.expect.GET("/api/v1/wallets/{id}/dashboard/summary").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	summary.Value("yearly_dividends").Number().IsEqual(0)
}
