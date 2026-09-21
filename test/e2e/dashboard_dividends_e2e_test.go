//go:build integration

package e2e

import "net/http"

func (s *E2ESuite) TestDashboard_Dividends_Success() {
	walletID := s.createWallet("Carteira Dividends")

	dividends := s.expect.GET("/api/v1/wallets/{id}/dashboard/dividends").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	dividends.Value("items").Array().Length().IsEqual(0)
}

func (s *E2ESuite) TestDashboard_Dividends_WalletNotFound() {
	s.expect.GET("/api/v1/wallets/{id}/dashboard/dividends").
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}
