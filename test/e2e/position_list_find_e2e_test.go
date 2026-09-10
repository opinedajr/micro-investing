//go:build integration

package e2e

import (
	"net/http"

	"github.com/opinedajr/micro-investing/internal/position"
)

type positionFixture struct {
	walletID string
	stockIDs map[string]string
}

func (s *E2ESuite) positionOrderingFixture() positionFixture {
	walletID := s.positionWalletID("Carteira Ordenação")

	petr4 := s.positionStockID("PETR4")
	vale3 := s.positionStockID("VALE3")
	mglu3 := s.positionStockID("MGLU3")

	s.positionInsertPrice(petr4, 7500)
	s.positionInsertPrice(vale3, 12000)
	s.positionInsertPrice(mglu3, 5000)

	s.positionCreate(walletID, petr4, 100, 5000)
	s.positionCreate(walletID, vale3, 100, 10000)
	s.positionCreate(walletID, mglu3, 10, 1000)

	return positionFixture{
		walletID: walletID,
		stockIDs: map[string]string{
			"PETR4": petr4,
			"VALE3": vale3,
			"MGLU3": mglu3,
		},
	}
}

func (s *E2ESuite) positionCreate(walletID, stockID string, quantity, averagePrice int64) string {
	resp := s.expect.POST("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		WithJSON(position.CreatePositionInput{
			StockID:      stockID,
			Quantity:     quantity,
			AveragePrice: averagePrice,
		}).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	return resp.Value("id").String().Raw()
}

func (s *E2ESuite) positionListStockIDs(walletID, ticker, sort string) []string {
	req := s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID)

	if ticker != "" {
		req = req.WithQuery("ticker", ticker)
	}
	if sort != "" {
		req = req.WithQuery("sort", sort)
	}

	arr := req.Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	ids := make([]string, 0, len(arr.Iter()))
	for _, item := range arr.Iter() {
		ids = append(ids, item.Object().Value("stock_id").String().Raw())
	}
	return ids
}

func (s *E2ESuite) TestPosition_List_Empty() {
	walletID := s.positionWalletID("Carteira Vazia")

	resp := s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	resp.Value("data").Array().IsEmpty()
}

func (s *E2ESuite) TestPosition_List_DefaultBalanceDesc() {
	f := s.positionOrderingFixture()

	got := s.positionListStockIDs(f.walletID, "", "")

	s.Require().Equal([]string{f.stockIDs["VALE3"], f.stockIDs["PETR4"], f.stockIDs["MGLU3"]}, got)
}

func (s *E2ESuite) TestPosition_List_SortTicker() {
	f := s.positionOrderingFixture()

	asc := s.positionListStockIDs(f.walletID, "", "ticker")
	s.Require().Equal([]string{f.stockIDs["MGLU3"], f.stockIDs["PETR4"], f.stockIDs["VALE3"]}, asc)

	desc := s.positionListStockIDs(f.walletID, "", "-ticker")
	s.Require().Equal([]string{f.stockIDs["VALE3"], f.stockIDs["PETR4"], f.stockIDs["MGLU3"]}, desc)
}

func (s *E2ESuite) TestPosition_List_SortRank() {
	walletID := s.positionWalletID("Carteira Rank")

	hapv3 := s.positionStockID("HAPV3")
	mglu3 := s.positionStockID("MGLU3")
	bbdc4 := s.positionStockID("BBDC4")

	s.positionCreate(walletID, hapv3, 10, 1000)
	s.positionCreate(walletID, mglu3, 10, 1000)
	s.positionCreate(walletID, bbdc4, 10, 1000)

	asc := s.positionListStockIDs(walletID, "", "rank")
	s.Require().Equal([]string{hapv3, mglu3, bbdc4}, asc)

	desc := s.positionListStockIDs(walletID, "", "-rank")
	s.Require().Equal([]string{bbdc4, mglu3, hapv3}, desc)
}

func (s *E2ESuite) TestPosition_List_SortInvested() {
	f := s.positionOrderingFixture()

	asc := s.positionListStockIDs(f.walletID, "", "invested")
	s.Require().Equal([]string{f.stockIDs["MGLU3"], f.stockIDs["PETR4"], f.stockIDs["VALE3"]}, asc)

	desc := s.positionListStockIDs(f.walletID, "", "-invested")
	s.Require().Equal([]string{f.stockIDs["VALE3"], f.stockIDs["PETR4"], f.stockIDs["MGLU3"]}, desc)
}

func (s *E2ESuite) TestPosition_List_SortVariationPercent() {
	f := s.positionOrderingFixture()

	asc := s.positionListStockIDs(f.walletID, "", "variation_percent")
	s.Require().Equal([]string{f.stockIDs["VALE3"], f.stockIDs["PETR4"], f.stockIDs["MGLU3"]}, asc)

	desc := s.positionListStockIDs(f.walletID, "", "-variation_percent")
	s.Require().Equal([]string{f.stockIDs["MGLU3"], f.stockIDs["PETR4"], f.stockIDs["VALE3"]}, desc)
}

func (s *E2ESuite) TestPosition_List_SortPortfolioPercent() {
	f := s.positionOrderingFixture()

	asc := s.positionListStockIDs(f.walletID, "", "portfolio_percent")
	s.Require().Equal([]string{f.stockIDs["MGLU3"], f.stockIDs["PETR4"], f.stockIDs["VALE3"]}, asc)

	desc := s.positionListStockIDs(f.walletID, "", "-portfolio_percent")
	s.Require().Equal([]string{f.stockIDs["VALE3"], f.stockIDs["PETR4"], f.stockIDs["MGLU3"]}, desc)
}

func (s *E2ESuite) TestPosition_List_SortInvalidFallbackToDefault() {
	f := s.positionOrderingFixture()

	got := s.positionListStockIDs(f.walletID, "", "sector")

	s.Require().Equal([]string{f.stockIDs["VALE3"], f.stockIDs["PETR4"], f.stockIDs["MGLU3"]}, got)
}

func (s *E2ESuite) TestPosition_List_FilterTickerCaseInsensitive() {
	f := s.positionOrderingFixture()

	got := s.positionListStockIDs(f.walletID, "petr", "")

	s.Require().Equal([]string{f.stockIDs["PETR4"]}, got)
}

func (s *E2ESuite) TestPosition_List_FilterTickerNoMatch() {
	f := s.positionOrderingFixture()

	got := s.positionListStockIDs(f.walletID, "ZERO", "")

	s.Require().Empty(got)
}

func (s *E2ESuite) TestPosition_List_WalletNotFound() {
	s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}

func (s *E2ESuite) TestPosition_Find_Success() {
	walletID := s.positionWalletID("Carteira Find")
	stockID := s.positionStockID("PETR4")
	s.positionInsertPrice(stockID, 7500)

	positionID := s.positionCreate(walletID, stockID, 100, 5000)

	resp := s.expect.GET("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", positionID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	resp.Value("id").String().IsEqual(positionID)
	resp.Value("wallet_id").String().IsEqual(walletID)
	resp.Value("stock_id").String().IsEqual(stockID)
	resp.Value("quantity").Number().IsEqual(100)
	resp.Value("average_price").Number().IsEqual(5000)
	resp.Value("current_price").Number().IsEqual(7500)
	resp.Value("invested").Number().IsEqual(500000)
	resp.Value("balance").Number().IsEqual(750000)
	resp.Value("variation_value").Number().IsEqual(250000)
	resp.Value("variation_percent").Number().IsEqual(50)
	resp.Value("portfolio_percent").Number().IsEqual(100)
	resp.Value("created_at").String().NotEmpty()
	resp.Value("updated_at").String().NotEmpty()
}

func (s *E2ESuite) TestPosition_Find_NotFound() {
	walletID := s.positionWalletID("Carteira Find NF")

	s.expect.GET("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", "123e4567-e89b-12d3-a456-426614174000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("POSITION_NOT_FOUND")
}

func (s *E2ESuite) TestPosition_Find_OtherWallet() {
	walletA := s.positionWalletID("Carteira A")
	walletB := s.positionWalletID("Carteira B")
	stockID := s.positionStockID("PETR4")

	positionID := s.positionCreate(walletA, stockID, 10, 1000)

	s.expect.GET("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletB).
		WithPath("positionId", positionID).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("POSITION_NOT_FOUND")
}

func (s *E2ESuite) TestPosition_Find_WalletNotFound() {
	s.expect.GET("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		WithPath("positionId", "123e4567-e89b-12d3-a456-426614174000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}
