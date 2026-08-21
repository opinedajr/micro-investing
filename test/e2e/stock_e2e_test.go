//go:build integration

package e2e

import (
	"net/http"
)

func (s *E2ESuite) TestStock_List() {
	resp := s.expect.GET("/api/v1/stocks").
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	arr := resp.Value("data").Array()
	arr.Length().IsEqual(29)
	arr.First().Object().Value("ticker").String().IsEqual("ABEV3")
	arr.Last().Object().Value("ticker").String().IsEqual("WEGE3")
}

func (s *E2ESuite) TestStock_List_ResponseShape() {
	resp := s.expect.GET("/api/v1/stocks").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Array().First().Object()

	resp.Value("id").String().NotEmpty()
	resp.Value("ticker").String().IsEqual("ABEV3")
	resp.Value("name").String().NotEmpty()
	resp.Value("sector").String().NotEmpty()
	resp.Value("rank").Number().IsEqual(9)
	resp.Value("website").String().NotEmpty()
	resp.Value("created_at").String().NotEmpty()
	resp.Value("updated_at").String().NotEmpty()
}

func (s *E2ESuite) TestStock_FindByTicker() {
	resp := s.expect.GET("/api/v1/stocks/{ticker}").
		WithPath("ticker", "PETR4").
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	resp.Value("id").String().NotEmpty()
	resp.Value("ticker").String().IsEqual("PETR4")
	resp.Value("name").String().IsEqual("Petrobras PN")
	resp.Value("sector").String().IsEqual("Petróleo, Gás e Biocombustíveis")
	resp.Value("rank").Number().IsEqual(10)
	resp.Value("website").String().IsEqual("https://petrobras.com.br")
	resp.Value("created_at").String().NotEmpty()
	resp.Value("updated_at").String().NotEmpty()
}

func (s *E2ESuite) TestStock_FindByTicker_NotFound() {
	s.expect.GET("/api/v1/stocks/{ticker}").
		WithPath("ticker", "INVALID").
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("STOCK_NOT_FOUND")
}
