//go:build integration

package e2e

import (
	"fmt"
	"net/http"
	"time"
)

func (s *E2ESuite) TestDividends_Create_YearLowerBoundary() {
	walletID := s.createWallet("Carteira Dividendos Ano 1900")

	resp := s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   1900,
			"amount": 1000,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	resp.Value("year").Number().IsEqual(1900)
	resp.Value("amount").Number().IsEqual(1000)
}

func (s *E2ESuite) TestDividends_Create_YearUpperBoundary() {
	walletID := s.createWallet("Carteira Dividendos Ano Limite")
	nextYear := time.Now().Year() + 1

	resp := s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   nextYear,
			"amount": 200000,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	resp.Value("id").String().NotEmpty()
	resp.Value("year").Number().IsEqual(nextYear)
	resp.Value("amount").Number().IsEqual(200000)
}

func (s *E2ESuite) TestDividends_Create_YearAboveUpperBoundary() {
	walletID := s.createWallet("Carteira Dividendos Ano Alem")
	farYear := time.Now().Year() + 2

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   farYear,
			"amount": 1000,
		}).
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestDividends_Create_MinimumAmount() {
	walletID := s.createWallet("Carteira Dividendos Amount Minimo")

	resp := s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   2020,
			"amount": 1,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	resp.Value("year").Number().IsEqual(2020)
	resp.Value("amount").Number().IsEqual(1)
}

func (s *E2ESuite) TestDividends_Create_MalformedJSON() {
	walletID := s.createWallet("Carteira Dividendos Malformed")

	body := fmt.Sprintf(`{"year": %d, "amount": }`, time.Now().Year())

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithText(body).
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestDividends_List_InvalidYearFormat() {
	walletID := s.createWallet("Carteira Dividendos Ano Invalido")

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{"year": 2025, "amount": 100000}).
		Expect().
		Status(http.StatusCreated)

	s.expect.GET("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithQuery("year", "abc").
		Expect().
		Status(http.StatusBadRequest).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
}

func (s *E2ESuite) TestDividends_List_IsolatedPerWallet() {
	walletA := s.createWallet("Carteira Dividendos Isolada A")
	walletB := s.createWallet("Carteira Dividendos Isolada B")

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletA).
		WithJSON(map[string]interface{}{"year": 2023, "amount": 30000}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletA).
		WithJSON(map[string]interface{}{"year": 2024, "amount": 40000}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletB).
		WithJSON(map[string]interface{}{"year": 2023, "amount": 99000}).
		Expect().
		Status(http.StatusCreated)

	itemsA := s.expect.GET("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletA).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object().Value("items").Array()

	itemsA.Length().IsEqual(2)
	itemsA.Element(0).Object().Value("year").Number().IsEqual(2024)
	itemsA.Element(1).Object().Value("year").Number().IsEqual(2023)
	itemsA.Element(1).Object().Value("amount").Number().IsEqual(30000)

	itemsB := s.expect.GET("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletB).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object().Value("items").Array()

	itemsB.Length().IsEqual(1)
	itemsB.Element(0).Object().Value("year").Number().IsEqual(2023)
	itemsB.Element(0).Object().Value("amount").Number().IsEqual(99000)
}

func (s *E2ESuite) TestDividends_CreateThenList_RoundTrip() {
	walletID := s.createWallet("Carteira Dividendos Roundtrip")

	created := s.expect.POST("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{"year": 2022, "amount": 123456}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	createdID := created.Value("id").String().NotEmpty().Raw()

	items := s.expect.GET("/api/v1/wallets/{walletId}/dividends").
		WithPath("walletId", walletID).
		WithQuery("year", "2022").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object().Value("items").Array()

	items.Length().IsEqual(1)
	item := items.Element(0).Object()
	item.Value("id").String().IsEqual(createdID)
	item.Value("year").Number().IsEqual(2022)
	item.Value("amount").Number().IsEqual(123456)
}
