//go:build integration

package e2e

import "net/http"

func (s *E2ESuite) createAsset(walletID, assetType, date, description string, amount int64) string {
	resp := s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"type":        assetType,
			"date":        date,
			"description": description,
			"amount":      amount,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	return resp.Value("id").String().Raw()
}

func (s *E2ESuite) TestAsset_Update() {
	walletID := s.createWallet("Carteira Update Asset")
	assetID := s.createAsset(walletID, "stocks", "2026-07-15T12:00:00Z", "PETR4 - Petrobras", 150000)

	resp := s.expect.PUT("/api/v1/wallets/{walletId}/assets/{id}").
		WithPath("walletId", walletID).
		WithPath("id", assetID).
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-15T12:00:00Z",
			"description": "PETR4 - Petrobras (ajustado)",
			"amount":      200000,
		}).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	resp.Value("id").String().IsEqual(assetID)
	resp.Value("wallet_id").String().IsEqual(walletID)
	resp.Value("type").String().IsEqual("stocks")
	resp.Value("date").String().IsEqual("2026-07-15T12:00:00Z")
	resp.Value("description").String().IsEqual("PETR4 - Petrobras (ajustado)")
	resp.Value("amount").Number().IsEqual(200000)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(1)
	arr.First().Object().Value("amount").Number().IsEqual(200000)
}

func (s *E2ESuite) TestAsset_Update_ChangeType_RecalculatesPatrimony() {
	walletID := s.createWallet("Carteira Update Type")
	assetID := s.createAsset(walletID, "stocks", "2026-07-15T12:00:00Z", "Acao Original", 150000)

	s.expect.PUT("/api/v1/wallets/{walletId}/assets/{id}").
		WithPath("walletId", walletID).
		WithPath("id", assetID).
		WithJSON(map[string]interface{}{
			"type":        "fiis",
			"date":        "2026-07-15T12:00:00Z",
			"description": "FII Convertido",
			"amount":      150000,
		}).
		Expect().
		Status(http.StatusOK)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(2)
	arr.First().Object().Value("type").String().IsEqual("fiis")
	arr.First().Object().Value("amount").Number().IsEqual(150000)
	arr.Last().Object().Value("type").String().IsEqual("stocks")
	arr.Last().Object().Value("amount").Number().IsEqual(0)
}

func (s *E2ESuite) TestAsset_Update_ChangeMonth_RecalculatesPatrimony() {
	walletID := s.createWallet("Carteira Update Month")
	assetID := s.createAsset(walletID, "stocks", "2026-07-15T12:00:00Z", "Acao Julho", 150000)

	s.expect.PUT("/api/v1/wallets/{walletId}/assets/{id}").
		WithPath("walletId", walletID).
		WithPath("id", assetID).
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-08-05T12:00:00Z",
			"description": "Acao Agosto",
			"amount":      180000,
		}).
		Expect().
		Status(http.StatusOK)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(2)
	arr.First().Object().Value("month").Number().IsEqual(8)
	arr.First().Object().Value("amount").Number().IsEqual(180000)
	arr.Last().Object().Value("month").Number().IsEqual(7)
	arr.Last().Object().Value("amount").Number().IsEqual(0)
}

func (s *E2ESuite) TestAsset_Update_NotFound() {
	walletID := s.createWallet("Carteira Update NotFound")

	s.expect.PUT("/api/v1/wallets/{walletId}/assets/{id}").
		WithPath("walletId", walletID).
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-15T12:00:00Z",
			"description": "Ativo Inexistente",
			"amount":      1000,
		}).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("ASSET_NOT_FOUND")
}

func (s *E2ESuite) TestAsset_Update_WalletNotFound() {
	s.expect.PUT("/api/v1/wallets/{walletId}/assets/{id}").
		WithPath("walletId", "00000000-0000-0000-0000-000000000000").
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-15T12:00:00Z",
			"description": "Carteira Inexistente",
			"amount":      1000,
		}).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}

func (s *E2ESuite) TestAsset_Update_InvalidType() {
	walletID := s.createWallet("Carteira Update BadType")
	assetID := s.createAsset(walletID, "stocks", "2026-07-15T12:00:00Z", "Acao", 1000)

	s.expect.PUT("/api/v1/wallets/{walletId}/assets/{id}").
		WithPath("walletId", walletID).
		WithPath("id", assetID).
		WithJSON(map[string]interface{}{
			"type":        "crypto",
			"date":        "2026-07-15T12:00:00Z",
			"description": "Tipo Invalido",
			"amount":      1000,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestAsset_Update_FutureDate() {
	walletID := s.createWallet("Carteira Update Future")
	assetID := s.createAsset(walletID, "stocks", "2026-07-15T12:00:00Z", "Acao", 1000)

	s.expect.PUT("/api/v1/wallets/{walletId}/assets/{id}").
		WithPath("walletId", walletID).
		WithPath("id", assetID).
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "2099-01-01T00:00:00Z",
			"description": "Data Futura",
			"amount":      1000,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestAsset_Update_InvalidDateFormat() {
	walletID := s.createWallet("Carteira Update BadDate")
	assetID := s.createAsset(walletID, "stocks", "2026-07-15T12:00:00Z", "Acao", 1000)

	s.expect.PUT("/api/v1/wallets/{walletId}/assets/{id}").
		WithPath("walletId", walletID).
		WithPath("id", assetID).
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "15/07/2026",
			"description": "Data Invalida",
			"amount":      1000,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestAsset_Update_MissingFields() {
	walletID := s.createWallet("Carteira Update Missing")
	assetID := s.createAsset(walletID, "stocks", "2026-07-15T12:00:00Z", "Acao", 1000)

	s.expect.PUT("/api/v1/wallets/{walletId}/assets/{id}").
		WithPath("walletId", walletID).
		WithPath("id", assetID).
		WithJSON(map[string]interface{}{
			"type": "stocks",
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestAsset_List() {
	walletID := s.createWallet("Carteira List Assets")

	s.createAsset(walletID, "stocks", "2026-07-15T12:00:00Z", "Acao A", 100000)
	s.createAsset(walletID, "fiis", "2026-08-01T12:00:00Z", "FII B", 50000)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(2)
	arr.First().Object().Value("date").String().IsEqual("2026-08-01T12:00:00Z")
	arr.Last().Object().Value("date").String().IsEqual("2026-07-15T12:00:00Z")
}

func (s *E2ESuite) TestAsset_List_Empty() {
	walletID := s.createWallet("Carteira List Empty")

	obj := s.expect.GET("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	obj.NotContainsKey("data")
}

func (s *E2ESuite) TestAsset_List_FilterByType() {
	walletID := s.createWallet("Carteira List Type")

	s.createAsset(walletID, "stocks", "2026-07-15T12:00:00Z", "Acao A", 100000)
	s.createAsset(walletID, "fiis", "2026-07-20T12:00:00Z", "FII B", 50000)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithQuery("type", "stocks").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(1)
	arr.First().Object().Value("type").String().IsEqual("stocks")
}

func (s *E2ESuite) TestAsset_List_FilterByStartDate() {
	walletID := s.createWallet("Carteira List Start")

	s.createAsset(walletID, "stocks", "2026-07-10T12:00:00Z", "Acao Julho", 100000)
	s.createAsset(walletID, "stocks", "2026-08-05T12:00:00Z", "Acao Agosto", 150000)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithQuery("start_date", "2026-08-01").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(1)
	arr.First().Object().Value("date").String().IsEqual("2026-08-05T12:00:00Z")
}

func (s *E2ESuite) TestAsset_List_FilterByEndDate() {
	walletID := s.createWallet("Carteira List End")

	s.createAsset(walletID, "stocks", "2026-07-10T12:00:00Z", "Acao Inicio", 100000)
	s.createAsset(walletID, "stocks", "2026-07-31T12:00:00Z", "Acao Limite", 150000)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithQuery("end_date", "2026-07-31").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(2)
}

func (s *E2ESuite) TestAsset_List_FilterByRange() {
	walletID := s.createWallet("Carteira List Range")

	s.createAsset(walletID, "stocks", "2026-07-10T12:00:00Z", "Acao Inicio", 100000)
	s.createAsset(walletID, "stocks", "2026-07-20T12:00:00Z", "Acao Meio", 120000)
	s.createAsset(walletID, "stocks", "2026-08-05T12:00:00Z", "Acao Fora", 150000)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithQuery("start_date", "2026-07-01").
		WithQuery("end_date", "2026-07-31").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(2)
}

func (s *E2ESuite) TestAsset_List_FilterByTypeAndRange() {
	walletID := s.createWallet("Carteira List Combo")

	s.createAsset(walletID, "stocks", "2026-07-10T12:00:00Z", "Acao Julho", 100000)
	s.createAsset(walletID, "stocks", "2026-08-05T12:00:00Z", "Acao Agosto", 150000)
	s.createAsset(walletID, "fiis", "2026-07-15T12:00:00Z", "FII Julho", 50000)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithQuery("type", "stocks").
		WithQuery("start_date", "2026-07-01").
		WithQuery("end_date", "2026-07-31").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(1)
	arr.First().Object().Value("type").String().IsEqual("stocks")
	arr.First().Object().Value("description").String().IsEqual("Acao Julho")
}

func (s *E2ESuite) TestAsset_List_InvalidStartDateFormat() {
	walletID := s.createWallet("Carteira List BadStart")

	s.expect.GET("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithQuery("start_date", "15/07/2026").
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestAsset_List_InvalidEndDateFormat() {
	walletID := s.createWallet("Carteira List BadEnd")

	s.expect.GET("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithQuery("end_date", "31-07-2026").
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestAsset_List_StartDateAfterEndDate() {
	walletID := s.createWallet("Carteira List BadRange")

	s.expect.GET("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithQuery("start_date", "2026-08-01").
		WithQuery("end_date", "2026-07-01").
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestAsset_List_WalletNotFound() {
	s.expect.GET("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", "00000000-0000-0000-0000-000000000000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}
