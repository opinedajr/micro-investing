package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/api"
	"github.com/stretchr/testify/assert"
)

type mockServiceForRoutes struct {
	createFunc func(ctx context.Context, input CreateWalletInput) (*WalletOutput, error)
}

func (m *mockServiceForRoutes) Create(ctx context.Context, input CreateWalletInput) (*WalletOutput, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, input)
	}
	return nil, nil
}

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - registers wallet routes under /api/v1", func(t *testing.T) {
		mockSvc := &mockServiceForRoutes{
			createFunc: func(ctx context.Context, input CreateWalletInput) (*WalletOutput, error) {
				return &WalletOutput{
					ID:          "wallet-id",
					Name:        input.Name,
					Description: input.Description,
					CreatedAt:   "2026-07-25T12:00:00Z",
					UpdatedAt:   "2026-07-25T12:00:00Z",
				}, nil
			},
		}

		handler := NewHandler(mockSvc)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler)

		body := map[string]interface{}{
			"name":        "Minha Carteira",
			"description": "Ações e FIIs",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/v1/wallets", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response api.Response[*WalletOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Minha Carteira", response.Data.Name)
	})
}
