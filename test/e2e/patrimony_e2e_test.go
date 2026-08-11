//go:build integration

package e2e

import "net/http"

func (s *E2ESuite) createWallet(name string) string {
	payload := map[string]interface{}{
		"name": name,
	}
	resp := s.expect.POST("/api/v1/wallets").
		WithJSON(payload).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()
	return resp.Value("id").String().Raw()
}

func (s *E2ESuite) TestPatrimony_Create() {
	walletID := s.createWallet("Carteira Principal")

	resp := s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  7,
			"type":   "stocks",
			"amount": 150000,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	resp.Value("id").String().NotEmpty()
	resp.Value("wallet_id").String().IsEqual(walletID)
	resp.Value("year").Number().IsEqual(2026)
	resp.Value("month").Number().IsEqual(7)
	resp.Value("type").String().IsEqual("stocks")
	resp.Value("amount").Number().IsEqual(150000)
	resp.Value("created_at").String().NotEmpty()
	resp.Value("updated_at").String().NotEmpty()
}

func (s *E2ESuite) TestPatrimony_Create_InvalidType() {
	walletID := s.createWallet("Carteira Type")

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  7,
			"type":   "invalid_type",
			"amount": 1000,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity)
}

func (s *E2ESuite) TestPatrimony_Create_MissingFields() {
	walletID := s.createWallet("Carteira Fields")

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year": 2026,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity)
}

func (s *E2ESuite) TestPatrimony_Create_InvalidYear() {
	walletID := s.createWallet("Carteira Year")

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   1999,
			"month":  1,
			"type":   "stocks",
			"amount": 1000,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity)
}

func (s *E2ESuite) TestPatrimony_Create_InvalidMonth() {
	walletID := s.createWallet("Carteira Month")

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  13,
			"type":   "stocks",
			"amount": 1000,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity)
}

func (s *E2ESuite) TestPatrimony_Create_NegativeAmount() {
	walletID := s.createWallet("Carteira Amount")

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  7,
			"type":   "stocks",
			"amount": -100,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity)
}

func (s *E2ESuite) TestPatrimony_Create_Duplicate() {
	walletID := s.createWallet("Carteira Dupe")

	payload := map[string]interface{}{
		"year":   2026,
		"month":  8,
		"type":   "fiis",
		"amount": 50000,
	}

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(payload).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(payload).
		Expect().
		Status(http.StatusConflict).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("PATRIMONY_ALREADY_EXISTS")
}

func (s *E2ESuite) TestPatrimony_Create_WalletNotFound() {
	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", "00000000-0000-0000-0000-000000000000").
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  7,
			"type":   "stocks",
			"amount": 1000,
		}).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}

func (s *E2ESuite) TestPatrimony_List() {
	walletID := s.createWallet("Carteira List")

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  3,
			"type":   "stocks",
			"amount": 100000,
		}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  3,
			"type":   "fiis",
			"amount": 50000,
		}).
		Expect().
		Status(http.StatusCreated)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(2)
}

func (s *E2ESuite) TestPatrimony_List_Empty() {
	walletID := s.createWallet("Carteira Empty")

	obj := s.expect.GET("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	obj.NotContainsKey("data")
}

func (s *E2ESuite) TestPatrimony_List_FilterByType() {
	walletID := s.createWallet("Carteira FilterType")

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  5,
			"type":   "stocks",
			"amount": 75000,
		}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  5,
			"type":   "fixed_income",
			"amount": 25000,
		}).
		Expect().
		Status(http.StatusCreated)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithQuery("type", "stocks").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(1)
	arr.First().Object().Value("type").String().IsEqual("stocks")
}

func (s *E2ESuite) TestPatrimony_List_FilterByYear() {
	walletID := s.createWallet("Carteira FilterYear")

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2025,
			"month":  12,
			"type":   "stocks",
			"amount": 90000,
		}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  1,
			"type":   "stocks",
			"amount": 110000,
		}).
		Expect().
		Status(http.StatusCreated)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithQuery("year", "2025").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(1)
	arr.First().Object().Value("year").Number().IsEqual(2025)
}

func (s *E2ESuite) TestPatrimony_List_FilterByMonth() {
	walletID := s.createWallet("Carteira FilterMonth")

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  6,
			"type":   "stocks",
			"amount": 80000,
		}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  7,
			"type":   "stocks",
			"amount": 85000,
		}).
		Expect().
		Status(http.StatusCreated)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithQuery("month", "6").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(1)
	arr.First().Object().Value("month").Number().IsEqual(6)
}

func (s *E2ESuite) TestPatrimony_List_WalletNotFound() {
	s.expect.GET("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", "00000000-0000-0000-0000-000000000000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}

func (s *E2ESuite) TestPatrimony_Update() {
	walletID := s.createWallet("Carteira Update")

	createResp := s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  10,
			"type":   "emergency_reserve",
			"amount": 30000,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	patrimonyID := createResp.Value("id").String().Raw()

	updateResp := s.expect.PUT("/api/v1/wallets/{walletId}/patrimonies/{id}").
		WithPath("walletId", walletID).
		WithPath("id", patrimonyID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  10,
			"type":   "emergency_reserve",
			"amount": 45000,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	updateResp.Value("id").String().IsEqual(patrimonyID)
	updateResp.Value("amount").Number().IsEqual(45000)
	updateResp.Value("type").String().IsEqual("emergency_reserve")
}

func (s *E2ESuite) TestPatrimony_Update_NotFound() {
	walletID := s.createWallet("Carteira UpdateNF")

	s.expect.PUT("/api/v1/wallets/{walletId}/patrimonies/{id}").
		WithPath("walletId", walletID).
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  11,
			"type":   "stocks",
			"amount": 5000,
		}).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("PATRIMONY_NOT_FOUND")
}

func (s *E2ESuite) TestPatrimony_Update_Duplicate() {
	walletID := s.createWallet("Carteira UpdateDupe")

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  4,
			"type":   "liquid_cash",
			"amount": 10000,
		}).
		Expect().
		Status(http.StatusCreated)

	createResp := s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  5,
			"type":   "liquid_cash",
			"amount": 20000,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	patrimonyID := createResp.Value("id").String().Raw()

	s.expect.PUT("/api/v1/wallets/{walletId}/patrimonies/{id}").
		WithPath("walletId", walletID).
		WithPath("id", patrimonyID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  4,
			"type":   "liquid_cash",
			"amount": 20000,
		}).
		Expect().
		Status(http.StatusConflict).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("PATRIMONY_ALREADY_EXISTS")
}

func (s *E2ESuite) TestPatrimony_Update_ValidationError() {
	walletID := s.createWallet("Carteira UpdateVal")

	createResp := s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  9,
			"type":   "fixed_income",
			"amount": 60000,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	patrimonyID := createResp.Value("id").String().Raw()

	s.expect.PUT("/api/v1/wallets/{walletId}/patrimonies/{id}").
		WithPath("walletId", walletID).
		WithPath("id", patrimonyID).
		WithJSON(map[string]interface{}{
			"year":   2026,
			"month":  0,
			"type":   "fixed_income",
			"amount": 60000,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity)
}
