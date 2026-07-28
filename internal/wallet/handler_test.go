package wallet

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/api"
	"github.com/stretchr/testify/assert"
)

func TestHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - creates wallet with valid input", func(t *testing.T) {
		mockSvc := newMockService(func(ctx context.Context, input CreateWalletInput) (*WalletOutput, error) {
			return &WalletOutput{
				ID:          "wallet-id",
				Name:        input.Name,
				Description: input.Description,
				CreatedAt:   "2026-07-25T12:00:00Z",
				UpdatedAt:   "2026-07-25T12:00:00Z",
			}, nil
		})

		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"name":        "Minha Carteira",
			"description": "Ações e FIIs",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response api.Response[*WalletOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "Minha Carteira", response.Data.Name)
		assert.Equal(t, "Ações e FIIs", *response.Data.Description)
	})

	t.Run("error - returns 422 for invalid input", func(t *testing.T) {
		mockSvc := newMockService(nil)
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"name": "AB",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 409 when name already exists", func(t *testing.T) {
		mockSvc := newMockService(func(ctx context.Context, input CreateWalletInput) (*WalletOutput, error) {
			return nil, ErrWalletNameAlreadyExists
		})

		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"name": "Duplicada",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "WALLET_NAME_ALREADY_EXISTS", response.Error.Code)
	})

	t.Run("error - returns 500 for internal server error", func(t *testing.T) {
		mockSvc := newMockService(func(ctx context.Context, input CreateWalletInput) (*WalletOutput, error) {
			return nil, errors.New("database connection failed")
		})

		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"name": "Test",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Create(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "INTERNAL_ERROR", response.Error.Code)
	})
}
