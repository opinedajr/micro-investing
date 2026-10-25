//go:build integration

package e2e

import (
	"net/http"
	"time"
)

func (s *E2ESuite) positionDeleteFixture() (walletID, petr4PositionID, vale3PositionID, petr4StockID, vale3StockID string) {
	walletID = s.positionWalletID("Carteira Delete")
	petr4StockID = s.positionStockID("PETR4")
	vale3StockID = s.positionStockID("VALE3")

	s.positionInsertPrice(petr4StockID, 7500)
	s.positionInsertPrice(vale3StockID, 12000)

	petr4PositionID = s.positionCreate(walletID, petr4StockID, 100, 5000)
	vale3PositionID = s.positionCreate(walletID, vale3StockID, 100, 10000)

	return walletID, petr4PositionID, vale3PositionID, petr4StockID, vale3StockID
}

func (s *E2ESuite) TestPosition_Delete_Success_NoContent() {
	walletID, petr4PositionID, _, _, _ := s.positionDeleteFixture()

	s.expect.DELETE("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", petr4PositionID).
		Expect().
		Status(http.StatusNoContent).
		Body().IsEmpty()
}

func (s *E2ESuite) TestPosition_Delete_RedistributesPortfolioPercent() {
	walletID, petr4PositionID, _, petr4StockID, vale3StockID := s.positionDeleteFixture()

	s.expect.DELETE("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", petr4PositionID).
		Expect().
		Status(http.StatusNoContent)

	arr := s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	s.Require().Len(arr.Iter(), 1)
	item := arr.First().Object()
	item.Value("stock_id").String().IsEqual(vale3StockID)
	item.Value("invested").Number().IsEqual(1000000)
	item.Value("portfolio_percent").Number().IsEqual(100)

	var deletedCount int64
	err := s.container.DB().
		Table("positions").
		Where("id = ?", petr4PositionID).
		Count(&deletedCount).Error
	s.Require().NoError(err)
	s.Require().Equal(int64(0), deletedCount)

	var petr4PercentCount int64
	err = s.container.DB().
		Table("positions").
		Where("wallet_id = ? AND stock_id = ?", walletID, petr4StockID).
		Count(&petr4PercentCount).Error
	s.Require().NoError(err)
	s.Require().Equal(int64(0), petr4PercentCount)
}

func (s *E2ESuite) TestPosition_Delete_LastPosition() {
	walletID := s.positionWalletID("Carteira Última Position")
	stockID := s.positionStockID("PETR4")
	positionID := s.positionCreate(walletID, stockID, 10, 1000)

	s.expect.DELETE("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", positionID).
		Expect().
		Status(http.StatusNoContent)

	s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array().
		Length().IsEqual(0)
}

func (s *E2ESuite) TestPosition_Delete_PositionNotFound() {
	walletID, _, _, _, _ := s.positionDeleteFixture()

	s.expect.DELETE("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", "123e4567-e89b-12d3-a456-426614174000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().
		Value("code").String().IsEqual("POSITION_NOT_FOUND")
}

func (s *E2ESuite) TestPosition_Delete_CrossWallet() {
	walletA, petr4PositionID, _, _, _ := s.positionDeleteFixture()
	walletB := s.positionWalletID("Carteira Delete B")

	s.expect.DELETE("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletB).
		WithPath("positionId", petr4PositionID).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().
		Value("code").String().IsEqual("POSITION_NOT_FOUND")

	s.expect.GET("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletA).
		WithPath("positionId", petr4PositionID).
		Expect().
		Status(http.StatusOK)

	s.Require().NotEqual(walletA, walletB)
}

func (s *E2ESuite) TestPosition_Delete_WalletNotFound() {
	s.expect.DELETE("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", "123e4567-e89b-12d3-a456-426614174000").
		WithPath("positionId", "123e4567-e89b-12d3-a456-426614174000").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().
		Value("code").String().IsEqual("WALLET_NOT_FOUND")
}

func (s *E2ESuite) TestPosition_Delete_Regression() {
	walletID, petr4PositionID, vale3PositionID, _, _ := s.positionDeleteFixture()

	s.expect.DELETE("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", petr4PositionID).
		Expect().
		Status(http.StatusNoContent)

	s.expect.GET("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", petr4PositionID).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().
		Value("code").String().IsEqual("POSITION_NOT_FOUND")

	list := s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	list.Length().IsEqual(1)
	list.First().Object().Value("id").String().IsEqual(vale3PositionID)

	newPositionID := s.positionCreate(walletID, s.positionStockID("ITUB4"), 5, 2000)
	s.Require().NotEmpty(newPositionID)

	arr := s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	s.Require().Len(arr.Iter(), 2)
	var sum float64
	for _, item := range arr.Iter() {
		sum += item.Object().Value("portfolio_percent").Number().Raw()
	}
	s.InDelta(100.0, sum, 0.001)

	var updatedPercent float64
	err := s.container.DB().
		Table("positions").
		Select("portfolio_percent").
		Where("id = ?", vale3PositionID).
		Scan(&updatedPercent).Error
	s.Require().NoError(err)
	s.InDelta(99.0099, updatedPercent, 0.001)

	s.expect.DELETE("/api/v1/wallets/{id}/positions/{positionId}").
		WithPath("id", walletID).
		WithPath("positionId", vale3PositionID).
		Expect().
		Status(http.StatusNoContent)

	remaining := s.expect.GET("/api/v1/wallets/{id}/positions").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array()

	remaining.Length().IsEqual(1)
	remaining.First().Object().Value("id").String().IsEqual(newPositionID)

	var remains int64
	err = s.container.DB().
		Table("positions").
		Where("wallet_id = ? AND updated_at > ?", walletID, time.Now().Add(-24*time.Hour)).
		Count(&remains).Error
	s.Require().NoError(err)
}
