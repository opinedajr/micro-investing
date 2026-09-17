//go:build integration

package e2e

import (
	"encoding/json"
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

func (s *E2ESuite) createPatrimony(walletID string, year, month int, assetType string, amount int64) {
	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{
			"year":   year,
			"month":  month,
			"type":   assetType,
			"amount": amount,
		}).
		Expect().
		Status(http.StatusCreated)
}

func (s *E2ESuite) TestDashboard_Allocation_Success() {
	walletID := s.createWallet("Carteira Allocation")

	s.createPatrimony(walletID, 2026, 3, "stocks", 500000)
	s.createPatrimony(walletID, 2026, 3, "fixed_income", 500000)
	s.createPatrimony(walletID, 2026, 3, "emergency_reserve", 250000)
	s.createPatrimony(walletID, 2026, 3, "liquid_cash", 250000)

	resp := s.expect.GET("/api/v1/wallets/{id}/dashboard/allocation").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK)

	var payload struct {
		Data struct {
			Items []struct {
				Type       string  `json:"type"`
				Amount     int64   `json:"amount"`
				Percentage float64 `json:"percentage"`
			} `json:"items"`
			Total int64 `json:"total"`
		} `json:"data"`
	}
	s.Require().NoError(json.Unmarshal([]byte(resp.Body().Raw()), &payload))

	s.Equal(int64(1500000), payload.Data.Total)
	s.Len(payload.Data.Items, 4)

	expectedAmounts := map[string]int64{
		"stocks":            500000,
		"fixed_income":      500000,
		"emergency_reserve": 250000,
		"liquid_cash":       250000,
	}
	for _, item := range payload.Data.Items {
		s.Equal(expectedAmounts[item.Type], item.Amount, "unexpected amount for %s", item.Type)
		s.InDelta(100.0*float64(item.Amount)/1500000.0, item.Percentage, 0.01)
	}
}

func (s *E2ESuite) TestDashboard_Allocation_EmptyWallet() {
	walletID := s.createWallet("Carteira Allocation Vazia")

	allocation := s.expect.GET("/api/v1/wallets/{id}/dashboard/allocation").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	allocation.Value("total").Number().IsEqual(0)
	allocation.Value("items").Array().Length().IsEqual(0)
}

func (s *E2ESuite) TestDashboard_Allocation_WalletNotFound() {
	s.expect.GET("/api/v1/wallets/{id}/dashboard/allocation").
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}

func (s *E2ESuite) TestDashboard_Evolution_YearQuarter() {
	walletID := s.createWallet("Carteira Evolution Trimestre")

	s.createPatrimony(walletID, 2026, 4, "stocks", 400000)
	s.createPatrimony(walletID, 2026, 4, "fixed_income", 500000)
	s.createPatrimony(walletID, 2026, 4, "emergency_reserve", 300000)

	s.createPatrimony(walletID, 2026, 5, "stocks", 500000)
	s.createPatrimony(walletID, 2026, 5, "fixed_income", 500000)
	s.createPatrimony(walletID, 2026, 5, "emergency_reserve", 350000)

	s.createPatrimony(walletID, 2026, 6, "stocks", 600000)
	s.createPatrimony(walletID, 2026, 6, "fixed_income", 550000)
	s.createPatrimony(walletID, 2026, 6, "emergency_reserve", 350000)

	evolution := s.expect.GET("/api/v1/wallets/{id}/dashboard/evolution").
		WithPath("id", walletID).
		WithQuery("year", 2026).
		WithQuery("quarter", 2).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	total := evolution.Value("total").Array()
	total.Length().IsEqual(3)
	total.Value(0).Object().Value("year").Number().IsEqual(2026)
	total.Value(0).Object().Value("month").Number().IsEqual(4)
	total.Value(0).Object().Value("amount").Number().IsEqual(1200000)
	total.Value(1).Object().Value("amount").Number().IsEqual(1350000)
	total.Value(2).Object().Value("amount").Number().IsEqual(1500000)

	byCategory := evolution.Value("by_category").Object()

	fixedIncome := byCategory.Value("fixed_income").Array()
	fixedIncome.Length().IsEqual(3)
	fixedIncome.Value(0).Object().Value("amount").Number().IsEqual(500000)
	fixedIncome.Value(1).Object().Value("amount").Number().IsEqual(500000)
	fixedIncome.Value(2).Object().Value("amount").Number().IsEqual(550000)

	stocks := byCategory.Value("stocks").Array()
	stocks.Length().IsEqual(3)
	stocks.Value(0).Object().Value("amount").Number().IsEqual(400000)
	stocks.Value(1).Object().Value("amount").Number().IsEqual(500000)
	stocks.Value(2).Object().Value("amount").Number().IsEqual(600000)

	emergencyReserve := byCategory.Value("emergency_reserve").Array()
	emergencyReserve.Length().IsEqual(3)
	emergencyReserve.Value(0).Object().Value("amount").Number().IsEqual(300000)
	emergencyReserve.Value(1).Object().Value("amount").Number().IsEqual(350000)
	emergencyReserve.Value(2).Object().Value("amount").Number().IsEqual(350000)
}

func (s *E2ESuite) TestDashboard_Evolution_YearOnly() {
	walletID := s.createWallet("Carteira Evolution Ano")

	s.createPatrimony(walletID, 2026, 3, "stocks", 200000)

	evolution := s.expect.GET("/api/v1/wallets/{id}/dashboard/evolution").
		WithPath("id", walletID).
		WithQuery("year", 2026).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	total := evolution.Value("total").Array()
	total.Length().IsEqual(12)
	total.Value(0).Object().Value("month").Number().IsEqual(1)
	total.Value(0).Object().Value("amount").Number().IsEqual(0)
	total.Value(2).Object().Value("month").Number().IsEqual(3)
	total.Value(2).Object().Value("amount").Number().IsEqual(200000)
	total.Value(11).Object().Value("month").Number().IsEqual(12)
	total.Value(11).Object().Value("amount").Number().IsEqual(200000)
}

func (s *E2ESuite) TestDashboard_Evolution_DefaultLast12Months() {
	walletID := s.createWallet("Carteira Evolution Default")

	evolution := s.expect.GET("/api/v1/wallets/{id}/dashboard/evolution").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	total := evolution.Value("total").Array()
	total.Length().IsEqual(12)
	total.Value(0).Object().Value("amount").Number().IsEqual(0)
	total.Value(11).Object().Value("amount").Number().IsEqual(0)

	byCategory := evolution.Value("by_category").Object()
	byCategory.Value("fixed_income").Array().Length().IsEqual(12)
	byCategory.Value("stocks").Array().Length().IsEqual(12)
	byCategory.Value("emergency_reserve").Array().Length().IsEqual(12)
}

func (s *E2ESuite) TestDashboard_Evolution_CarryForward() {
	walletID := s.createWallet("Carteira Evolution CarryForward")

	s.createPatrimony(walletID, 2026, 4, "stocks", 400000)
	s.createPatrimony(walletID, 2026, 4, "fixed_income", 500000)

	s.createPatrimony(walletID, 2026, 6, "stocks", 600000)
	s.createPatrimony(walletID, 2026, 6, "fixed_income", 550000)

	evolution := s.expect.GET("/api/v1/wallets/{id}/dashboard/evolution").
		WithPath("id", walletID).
		WithQuery("year", 2026).
		WithQuery("quarter", 2).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	total := evolution.Value("total").Array()
	total.Length().IsEqual(3)
	total.Value(0).Object().Value("amount").Number().IsEqual(900000)
	total.Value(1).Object().Value("amount").Number().IsEqual(900000)
	total.Value(2).Object().Value("amount").Number().IsEqual(1150000)
}

func (s *E2ESuite) TestDashboard_Evolution_CarryForwardPerCategory() {
	walletID := s.createWallet("Carteira Evolution Parcial")

	s.createPatrimony(walletID, 2026, 4, "stocks", 400000)
	s.createPatrimony(walletID, 2026, 4, "fixed_income", 500000)
	s.createPatrimony(walletID, 2026, 4, "emergency_reserve", 300000)

	s.createPatrimony(walletID, 2026, 5, "stocks", 500000)
	s.createPatrimony(walletID, 2026, 5, "fixed_income", 500000)

	s.createPatrimony(walletID, 2026, 6, "fixed_income", 550000)

	evolution := s.expect.GET("/api/v1/wallets/{id}/dashboard/evolution").
		WithPath("id", walletID).
		WithQuery("year", 2026).
		WithQuery("quarter", 2).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	total := evolution.Value("total").Array()
	total.Value(0).Object().Value("amount").Number().IsEqual(1200000)
	total.Value(1).Object().Value("amount").Number().IsEqual(1000000)
	total.Value(2).Object().Value("amount").Number().IsEqual(550000)

	byCategory := evolution.Value("by_category").Object()

	stocks := byCategory.Value("stocks").Array()
	stocks.Value(0).Object().Value("amount").Number().IsEqual(400000)
	stocks.Value(1).Object().Value("amount").Number().IsEqual(500000)
	stocks.Value(2).Object().Value("amount").Number().IsEqual(500000)

	fixedIncome := byCategory.Value("fixed_income").Array()
	fixedIncome.Value(0).Object().Value("amount").Number().IsEqual(500000)
	fixedIncome.Value(1).Object().Value("amount").Number().IsEqual(500000)
	fixedIncome.Value(2).Object().Value("amount").Number().IsEqual(550000)

	emergencyReserve := byCategory.Value("emergency_reserve").Array()
	emergencyReserve.Value(0).Object().Value("amount").Number().IsEqual(300000)
	emergencyReserve.Value(1).Object().Value("amount").Number().IsEqual(300000)
	emergencyReserve.Value(2).Object().Value("amount").Number().IsEqual(300000)
}

func (s *E2ESuite) TestDashboard_Evolution_TotalIncludesAllCategories() {
	walletID := s.createWallet("Carteira Evolution Total")

	s.createPatrimony(walletID, 2026, 4, "stocks", 400000)
	s.createPatrimony(walletID, 2026, 4, "fiis", 100000)
	s.createPatrimony(walletID, 2026, 4, "fixed_income", 500000)
	s.createPatrimony(walletID, 2026, 4, "emergency_reserve", 300000)
	s.createPatrimony(walletID, 2026, 4, "liquid_cash", 50000)

	evolution := s.expect.GET("/api/v1/wallets/{id}/dashboard/evolution").
		WithPath("id", walletID).
		WithQuery("year", 2026).
		WithQuery("quarter", 2).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	total := evolution.Value("total").Array()
	total.Value(0).Object().Value("amount").Number().IsEqual(1350000)

	byCategory := evolution.Value("by_category").Object()
	byCategory.Value("fixed_income").Array().Value(0).Object().Value("amount").Number().IsEqual(500000)
	byCategory.Value("stocks").Array().Value(0).Object().Value("amount").Number().IsEqual(400000)
	byCategory.Value("emergency_reserve").Array().Value(0).Object().Value("amount").Number().IsEqual(300000)
	byCategory.NotContainsKey("fiis")
	byCategory.NotContainsKey("liquid_cash")
}

func (s *E2ESuite) TestDashboard_Evolution_ValidationErrors() {
	walletID := s.createWallet("Carteira Evolution Validacao")

	invalidQueries := []string{
		"quarter=5",
		"quarter=0",
		"quarter=-1",
		"year=2026&quarter=5",
		"year=2026&quarter=0",
		"year=2026&quarter=-1",
		"year=2026&quarter=abc",
		"year=abc&quarter=2",
		"year=0&quarter=2",
		"year=-2026&quarter=2",
		"quarter=2",
	}

	for _, qs := range invalidQueries {
		s.expect.GET("/api/v1/wallets/{id}/dashboard/evolution").
			WithPath("id", walletID).
			WithQueryString(qs).
			Expect().
			Status(http.StatusBadRequest).
			JSON().Object().Value("error").Object().Value("code").String().IsEqual("VALIDATION_ERROR")
	}
}

func (s *E2ESuite) TestDashboard_Evolution_WalletNotFound() {
	s.expect.GET("/api/v1/wallets/{id}/dashboard/evolution").
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}
