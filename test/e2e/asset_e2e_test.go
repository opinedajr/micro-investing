//go:build integration

package e2e

import (
	"net/http"
	"strings"
)

func (s *E2ESuite) TestAsset_Create() {
	walletID := s.createWallet("Carteira Assets")

	resp := s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-15T12:00:00Z",
			"description": "PETR4 - Petrobras",
			"amount":      150000,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	resp.Value("id").String().NotEmpty()
	resp.Value("wallet_id").String().IsEqual(walletID)
	resp.Value("type").String().IsEqual("stocks")
	resp.Value("date").String().IsEqual("2026-07-15T12:00:00Z")
	resp.Value("description").String().IsEqual("PETR4 - Petrobras")
	resp.Value("amount").Number().IsEqual(150000)
	resp.Value("created_at").String().NotEmpty()
	resp.Value("updated_at").String().NotEmpty()
}

func (s *E2ESuite) TestAsset_Create_RecalculatesPatrimony() {
	walletID := s.createWallet("Carteira Recalc")

	s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-15T12:00:00Z",
			"description": "PETR4 - Petrobras",
			"amount":      150000,
		}).
		Expect().
		Status(http.StatusCreated)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(1)
	arr.First().Object().Value("type").String().IsEqual("stocks")
	arr.First().Object().Value("year").Number().IsEqual(2026)
	arr.First().Object().Value("month").Number().IsEqual(7)
	arr.First().Object().Value("amount").Number().IsEqual(150000)
}

func (s *E2ESuite) TestAsset_Create_AccumulatesPatrimony() {
	walletID := s.createWallet("Carteira Accum")

	s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"type":        "fiis",
			"date":        "2026-07-20T12:00:00Z",
			"description": "FII Shopping",
			"amount":      50000,
		}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"type":        "fiis",
			"date":        "2026-07-25T12:00:00Z",
			"description": "FII Logistica",
			"amount":      25000,
		}).
		Expect().
		Status(http.StatusCreated)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(1)
	arr.First().Object().Value("type").String().IsEqual("fiis")
	arr.First().Object().Value("amount").Number().IsEqual(75000)
}

func (s *E2ESuite) TestAsset_Delete() {
	walletID := s.createWallet("Carteira Delete")

	createResp := s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-15T12:00:00Z",
			"description": "PETR4 - Petrobras",
			"amount":      150000,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	assetID := createResp.Value("id").String().Raw()

	s.expect.DELETE("/api/v1/wallets/{walletId}/assets/{id}").
		WithPath("walletId", walletID).
		WithPath("id", assetID).
		Expect().
		Status(http.StatusNoContent)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(1)
	arr.First().Object().Value("amount").Number().IsEqual(0)
}

func (s *E2ESuite) TestAsset_Delete_RecalculatesRemaining() {
	walletID := s.createWallet("Carteira DelRecalc")

	createResp := s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-15T12:00:00Z",
			"description": "Acao Um",
			"amount":      100000,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	assetOneID := createResp.Value("id").String().Raw()

	s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-15T12:00:00Z",
			"description": "Acao Dois",
			"amount":      50000,
		}).
		Expect().
		Status(http.StatusCreated)

	s.expect.DELETE("/api/v1/wallets/{walletId}/assets/{id}").
		WithPath("walletId", walletID).
		WithPath("id", assetOneID).
		Expect().
		Status(http.StatusNoContent)

	arr := s.expect.GET("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	arr.Length().IsEqual(1)
	arr.First().Object().Value("amount").Number().IsEqual(50000)
}

func (s *E2ESuite) TestAsset_Create_InvalidType() {
	walletID := s.createWallet("Carteira AssetType")

	s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"type":        "crypto",
			"date":        "2026-07-15T12:00:00Z",
			"description": "Bitcoin",
			"amount":      1000,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestAsset_Create_FutureDate() {
	walletID := s.createWallet("Carteira FutDate")

	s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
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

func (s *E2ESuite) TestAsset_Create_ShortDescription() {
	walletID := s.createWallet("Carteira ShortDesc")

	s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-15T12:00:00Z",
			"description": "ab",
			"amount":      1000,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestAsset_Create_LongDescription() {
	walletID := s.createWallet("Carteira LongDesc")

	s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-15T12:00:00Z",
			"description": strings.Repeat("a", 101),
			"amount":      1000,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestAsset_Create_ZeroAmount() {
	walletID := s.createWallet("Carteira Zero")

	s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-15T12:00:00Z",
			"description": "Valor Zero",
			"amount":      0,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestAsset_Create_NegativeAmount() {
	walletID := s.createWallet("Carteira Neg")

	s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"type":        "stocks",
			"date":        "2026-07-15T12:00:00Z",
			"description": "Valor Negativo",
			"amount":      -100,
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestAsset_Create_InvalidDateFormat() {
	walletID := s.createWallet("Carteira BadDate")

	s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
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

func (s *E2ESuite) TestAsset_Create_MissingFields() {
	walletID := s.createWallet("Carteira Missing")

	s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"type": "stocks",
		}).
		Expect().
		Status(http.StatusUnprocessableEntity).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestAsset_Create_WalletNotFound() {
	s.expect.POST("/api/v1/wallets/{walletId}/assets").
		WithPath("walletId", "00000000-0000-0000-0000-000000000000").
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

func (s *E2ESuite) TestAsset_Delete_NotFound() {
	walletID := s.createWallet("Carteira DelNF")

	s.expect.DELETE("/api/v1/wallets/{walletId}/assets/{id}").
		WithPath("walletId", walletID).
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("ASSET_NOT_FOUND")
}
