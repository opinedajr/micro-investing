//go:build integration

package e2e

import (
	"net/http"

	"github.com/opinedajr/micro-investing/internal/wallet"
)

func (s *E2ESuite) TestWallet_CreateAndGet() {
	payload := wallet.CreateWalletInput{
		Name:        "Reserva de Emergência",
		Description: strPtr("Minha carteira de liquidez diária"),
	}

	resp := s.expect.POST("/api/v1/wallets").
		WithJSON(payload).
		Expect().
		Status(http.StatusCreated).
		JSON().Object().Value("data").Object()

	resp.Value("id").String().NotEmpty()
	resp.Value("name").String().IsEqual("Reserva de Emergência")
	resp.Value("description").String().IsEqual("Minha carteira de liquidez diária")
	resp.Value("created_at").String().NotEmpty()
	resp.Value("updated_at").String().NotEmpty()

	walletID := resp.Value("id").String().Raw()

	getResp := s.expect.GET("/api/v1/wallets/{id}").
		WithPath("id", walletID).
		Expect().
		Status(http.StatusOK).
		JSON().Object().Value("data").Object()

	getResp.Value("id").String().IsEqual(walletID)
	getResp.Value("name").String().IsEqual("Reserva de Emergência")
}

func (s *E2ESuite) TestWallet_Create_InvalidPayload() {
	payload := wallet.CreateWalletInput{
		Name: "",
	}

	s.expect.POST("/api/v1/wallets").
		WithJSON(payload).
		Expect().
		Status(http.StatusUnprocessableEntity)
}

func (s *E2ESuite) TestWallet_Create_Conflict() {
	payload := wallet.CreateWalletInput{
		Name:        "Carteira Duplicada",
		Description: strPtr("Teste de conflito"),
	}

	s.expect.POST("/api/v1/wallets").
		WithJSON(payload).
		Expect().
		Status(http.StatusCreated)

	s.expect.POST("/api/v1/wallets").
		WithJSON(payload).
		Expect().
		Status(http.StatusConflict).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NAME_ALREADY_EXISTS")
}

func (s *E2ESuite) TestWallet_Find_NotFound() {
	randomUUID := "123e4567-e89b-12d3-a456-426614174000"

	s.expect.GET("/api/v1/wallets/{id}").
		WithPath("id", randomUUID).
		Expect().
		Status(http.StatusNotFound).
		JSON().Object().Value("error").Object().Value("code").String().IsEqual("WALLET_NOT_FOUND")
}

func (s *E2ESuite) TestWallet_List() {
	payload1 := wallet.CreateWalletInput{
		Name: "Carteira Fixa",
	}
	s.expect.POST("/api/v1/wallets").
		WithJSON(payload1).
		Expect().
		Status(http.StatusCreated)

	payload2 := wallet.CreateWalletInput{
		Name: "Carteira Variável",
	}
	s.expect.POST("/api/v1/wallets").
		WithJSON(payload2).
		Expect().
		Status(http.StatusCreated)

	resp := s.expect.GET("/api/v1/wallets").
		Expect().
		Status(http.StatusOK).
		JSON().Object()

	arr := resp.Value("data").Array()
	arr.Length().IsEqual(2)

	arr.First().Object().Value("name").String().IsEqual("Carteira Fixa")
	arr.Last().Object().Value("name").String().IsEqual("Carteira Variável")
}

func strPtr(s string) *string {
	return &s
}
