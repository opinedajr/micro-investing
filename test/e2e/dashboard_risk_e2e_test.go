//go:build integration

package e2e

import (
	"net/http"
	"time"

	"github.com/opinedajr/micro-investing/internal/dashboard"
	"github.com/opinedajr/micro-investing/internal/shared/api"
)

func (s *E2ESuite) TestDashboard_Risk_Success() {
	walletID := s.createWallet("Carteira Risk Multiplos Ranks")

	petr4ID := s.positionStockID("PETR4")
	vale3ID := s.positionStockID("VALE3")
	hapv3ID := s.positionStockID("HAPV3")
	prio3ID := s.positionStockID("PRIO3")

	s.positionCreate(walletID, petr4ID, 10, 10000)
	s.positionCreate(walletID, vale3ID, 10, 10000)
	s.positionCreate(walletID, hapv3ID, 10, 10000)
	s.positionCreate(walletID, prio3ID, 10, 10000)

	data := s.expect.GET("/api/v1/wallets/{id}/dashboard/risk").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	data.ContainsKey("items")
	data.ContainsKey("total")
	data.Value("total").Number().IsEqual(400000)
	data.Value("items").Array().Length().IsEqual(3)

	var resp api.Response[dashboard.RiskOutput]
	s.expect.GET("/api/v1/wallets/{id}/dashboard/risk").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Decode(&resp)

	s.Require().NotNil(resp.Data)
	s.Require().Equal(int64(400000), resp.Data.Total)
	s.Require().Len(resp.Data.Items, 3)

	amountByRank := map[int8]int64{}
	percentageByRank := map[int8]float64{}
	for _, item := range resp.Data.Items {
		amountByRank[item.Rank] = item.Amount
		percentageByRank[item.Rank] = item.Percentage
	}

	s.Require().Equal(int64(200000), amountByRank[10])
	s.Require().Equal(int64(100000), amountByRank[7])
	s.Require().Equal(int64(100000), amountByRank[5])
	s.InDelta(50.0, percentageByRank[10], 0.01)
	s.InDelta(25.0, percentageByRank[7], 0.01)
	s.InDelta(25.0, percentageByRank[5], 0.01)
}

func (s *E2ESuite) TestDashboard_Risk_EmptyWallet() {
	walletID := s.createWallet("Carteira Risk Vazia")

	data := s.expect.GET("/api/v1/wallets/{id}/dashboard/risk").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	data.Value("total").Number().IsEqual(0)
	data.Value("items").Array().Length().IsEqual(0)
}

func (s *E2ESuite) TestDashboard_Risk_SinglePosition() {
	walletID := s.createWallet("Carteira Risk Single")

	petr4ID := s.positionStockID("PETR4")
	s.positionCreate(walletID, petr4ID, 5, 10000)

	var resp api.Response[dashboard.RiskOutput]
	s.expect.GET("/api/v1/wallets/{id}/dashboard/risk").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Decode(&resp)

	s.Require().NotNil(resp.Data)
	s.Require().Equal(int64(50000), resp.Data.Total)
	s.Require().Len(resp.Data.Items, 1)

	item := resp.Data.Items[0]
	s.Require().Equal(int8(10), item.Rank)
	s.Require().Equal(int64(50000), item.Amount)
	s.InDelta(100.0, item.Percentage, 0.01)
}

func (s *E2ESuite) TestDashboard_Risk_WalletNotFound() {
	s.expect.GET("/api/v1/wallets/{id}/dashboard/risk").
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}

func (s *E2ESuite) TestDashboard_Risk_IgnoresOrphanPositions() {
	walletID := s.createWallet("Carteira Risk Orfa")

	petr4ID := s.positionStockID("PETR4")
	s.positionCreate(walletID, petr4ID, 10, 5000)

	now := time.Now()
	err := s.container.DB().Exec(
		`INSERT INTO positions (id, wallet_id, stock_id, quantity, average_price, current_price, invested, balance, variation_value, variation_percent, portfolio_percent, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		"orphan-position-1", walletID, "non-existent-stock-id", 10, 9999, 0, 99990, 0, 0, 0, 0, now, now,
	).Error
	s.Require().NoError(err)

	var resp api.Response[dashboard.RiskOutput]
	s.expect.GET("/api/v1/wallets/{id}/dashboard/risk").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Decode(&resp)

	s.Require().NotNil(resp.Data)
	s.Require().Equal(int64(50000), resp.Data.Total)
	s.Require().Len(resp.Data.Items, 1)

	item := resp.Data.Items[0]
	s.Require().Equal(int8(10), item.Rank)
	s.Require().Equal(int64(50000), item.Amount)
	s.InDelta(100.0, item.Percentage, 0.01)
}

func (s *E2ESuite) TestDashboard_Allocation_Regression() {
	walletID := s.createWallet("Carteira Allocation Regression")

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{"year": 2026, "month": 3, "type": "stocks", "amount": 200000}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{"year": 2026, "month": 3, "type": "fixed_income", "amount": 150000}).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets/{walletId}/patrimonies").
		WithPath("walletId", walletID).
		WithJSON(map[string]interface{}{"year": 2026, "month": 3, "type": "emergency_reserve", "amount": 50000}).
		Expect().
		Status(http.StatusCreated)

	var resp api.Response[dashboard.AllocationOutput]
	s.expect.GET("/api/v1/wallets/{id}/dashboard/allocation").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Decode(&resp)

	s.Require().NotNil(resp.Data)
	s.Require().Equal(int64(400000), resp.Data.Total)
	s.Require().Len(resp.Data.Items, 3)

	amountByType := map[string]int64{}
	percentageByType := map[string]float64{}
	for _, item := range resp.Data.Items {
		amountByType[item.Type] = item.Amount
		percentageByType[item.Type] = item.Percentage
	}

	s.Require().Equal(int64(200000), amountByType["stocks"])
	s.Require().Equal(int64(150000), amountByType["fixed_income"])
	s.Require().Equal(int64(50000), amountByType["emergency_reserve"])
	s.InDelta(50.0, percentageByType["stocks"], 0.01)
	s.InDelta(37.5, percentageByType["fixed_income"], 0.01)
	s.InDelta(12.5, percentageByType["emergency_reserve"], 0.01)
}
