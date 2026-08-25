package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/api"
	"github.com/opinedajr/micro-investing/internal/wallet"
	"github.com/stretchr/testify/assert"
)

type mockWalletService struct {
	findFunc func(ctx context.Context, id string) (*wallet.WalletOutput, error)
}

func (m *mockWalletService) Create(ctx context.Context, input wallet.CreateWalletInput) (*wallet.WalletOutput, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWalletService) List(ctx context.Context, userID string) ([]wallet.WalletOutput, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWalletService) Find(ctx context.Context, id string) (*wallet.WalletOutput, error) {
	if m.findFunc != nil {
		return m.findFunc(ctx, id)
	}
	return nil, errors.New("not found")
}

func (m *mockWalletService) Update(ctx context.Context, id string, input wallet.UpdateWalletInput) (*wallet.WalletOutput, error) {
	return nil, errors.New("not implemented")
}

func (m *mockWalletService) Delete(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func TestWalletMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - calls next when wallet exists", func(t *testing.T) {
		service := &mockWalletService{
			findFunc: func(ctx context.Context, id string) (*wallet.WalletOutput, error) {
				return &wallet.WalletOutput{ID: id}, nil
			},
		}

		r := gin.New()
		r.Use(WalletMiddleware(service))
		r.GET("/wallets/:id/patrimonies", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/wallets/wallet-id/patrimonies", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("error - returns 404 when wallet not found", func(t *testing.T) {
		service := &mockWalletService{
			findFunc: func(ctx context.Context, id string) (*wallet.WalletOutput, error) {
				return nil, errors.New("not found")
			},
		}

		r := gin.New()
		r.Use(WalletMiddleware(service))
		r.GET("/wallets/:id/patrimonies", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/wallets/non-existent/patrimonies", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "WALLET_NOT_FOUND", response.Error.Code)
	})

	t.Run("error - returns 404 when wallet id is missing", func(t *testing.T) {
		service := &mockWalletService{}

		r := gin.New()
		r.Use(WalletMiddleware(service))
		r.GET("/patrimonies", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/patrimonies", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
