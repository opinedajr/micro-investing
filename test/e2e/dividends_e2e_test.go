//go:build integration

package e2e

import "net/http"

func (s *E2ESuite) TestDividends_Create() {
	walletID := s.createWallet("Carteira Dividendos")

	resp := s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2025,
			"amount": 150000,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	resp.Value("id").String().NotEmpty()
	resp.Value("year").Number().IsEqual(2025)
	resp.Value("amount").Number().IsEqual(150000)
}

func (s *E2ESuite) TestDividends_Create_Duplicate() {
	walletID := s.createWallet("Carteira Dividendos Dupe")

	payload := map[string]interface{}{
		"year":   2025,
		"amount": 150000,
	}

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(payload).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(payload).
		Expect().
		Status(http.StatusConflict).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("DIVIDEND_ALREADY_EXISTS")
}

func (s *E2ESuite) TestDividends_Create_SameYearOtherWallet() {
	walletA := s.createWallet("Carteira Dividendos A")
	walletB := s.createWallet("Carteira Dividendos B")

	payload := map[string]interface{}{
		"year":   2025,
		"amount": 150000,
	}

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletA).
		WithJSON(payload).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletB).
		WithJSON(payload).
		Expect().
		Status(http.StatusCreated)
}

func (s *E2ESuite) TestDividends_Create_InvalidYear() {
	walletID := s.createWallet("Carteira Dividendos Year")

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   1899,
			"amount": 150000,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestDividends_Create_YearTooFarInFuture() {
	walletID := s.createWallet("Carteira Dividendos Future")

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2200,
			"amount": 150000,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestDividends_Create_NonPositiveAmount() {
	walletID := s.createWallet("Carteira Dividendos Amount")

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2025,
			"amount": 0,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2025,
			"amount": -100,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestDividends_Create_MissingFields() {
	walletID := s.createWallet("Carteira Dividendos Fields")

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year": 2025,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestDividends_Create_WalletNotFound() {
	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", "00000000-0000-0000-0000-000000000000").
		WithJSON(map[string]interface{}{
			"year":   2025,
			"amount": 150000,
		}).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}

func (s *E2ESuite) TestDividends_List() {
	walletID := s.createWallet("Carteira Dividendos List")

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{"year": 2024, "amount": 50000}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{"year": 2026, "amount": 150000}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{"year": 2025, "amount": 100000}).
		Expect().
		Status(http.StatusCreated)

	items := s.expect.GET("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object().Value("items").Array()

	items.Length().IsEqual(3)
	items.Element(0).Object().Value("year").Number().IsEqual(2026)
	items.Element(1).Object().Value("year").Number().IsEqual(2025)
	items.Element(2).Object().Value("year").Number().IsEqual(2024)
}

func (s *E2ESuite) TestDividends_List_Empty() {
	walletID := s.createWallet("Carteira Dividendos Empty")

	items := s.expect.GET("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object().Value("items").Array()

	items.IsEmpty()
}

func (s *E2ESuite) TestDividends_List_FilterByYear() {
	walletID := s.createWallet("Carteira Dividendos Filter")

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{"year": 2025, "amount": 100000}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{"year": 2026, "amount": 150000}).
		Expect().
		Status(http.StatusCreated)

	items := s.expect.GET("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithQuery("year", "2025").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object().Value("items").Array()

	items.Length().IsEqual(1)
	items.First().Object().Value("year").Number().IsEqual(2025)
}

func (s *E2ESuite) TestDividends_List_FilterWithoutMatch() {
	walletID := s.createWallet("Carteira Dividendos FilterEmpty")

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{"year": 2025, "amount": 100000}).
		Expect().
		Status(http.StatusCreated)

	items := s.expect.GET("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithQuery("year", "1990").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object().Value("items").Array()

	items.IsEmpty()
}

func (s *E2ESuite) TestDividends_List_WalletNotFound() {
	s.expect.GET("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", "00000000-0000-0000-0000-000000000000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}
