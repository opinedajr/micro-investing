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

func TestHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - lists all wallets", func(t *testing.T) {
		mockSvc := newMockServiceWithListAndFind(
			func(ctx context.Context, input CreateWalletInput) (*WalletOutput, error) {
				return nil, nil
			},
			func(ctx context.Context, userID string) ([]WalletOutput, error) {
				desc1 := "Ações"
				desc2 := "FIIs"
				return []WalletOutput{
					{
						ID:          "wallet-1",
						Name:        "Carteira Ações",
						Description: &desc1,
						CreatedAt:   "2026-07-25T10:00:00Z",
						UpdatedAt:   "2026-07-25T10:00:00Z",
					},
					{
						ID:          "wallet-2",
						Name:        "Carteira FIIs",
						Description: &desc2,
						CreatedAt:   "2026-07-25T11:00:00Z",
						UpdatedAt:   "2026-07-25T11:00:00Z",
					},
				}, nil
			},
			nil,
		)

		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/wallets", nil)

		handler.List(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[[]WalletOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 2)
		assert.Equal(t, "wallet-1", response.Data[0].ID)
		assert.Equal(t, "Carteira Ações", response.Data[0].Name)
		assert.Equal(t, "Ações", *response.Data[0].Description)
		assert.Nil(t, response.Meta)
	})
}

func TestHandler_Find(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - finds wallet by id", func(t *testing.T) {
		desc := "Minha Carteira de Ações"
		mockSvc := newMockServiceWithListAndFind(
			func(ctx context.Context, input CreateWalletInput) (*WalletOutput, error) {
				return nil, nil
			},
			nil,
			func(ctx context.Context, id string) (*WalletOutput, error) {
				return &WalletOutput{
					ID:          "wallet-123",
					Name:        "Carteira Principal",
					Description: &desc,
					CreatedAt:   "2026-07-25T12:00:00Z",
					UpdatedAt:   "2026-07-25T12:00:00Z",
				}, nil
			},
		)

		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/wallets/wallet-123", nil)
		c.Params = []gin.Param{{Key: "id", Value: "wallet-123"}}

		handler.Find(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[*WalletOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "wallet-123", response.Data.ID)
		assert.Equal(t, "Carteira Principal", response.Data.Name)
		assert.Equal(t, "Minha Carteira de Ações", *response.Data.Description)
	})

	t.Run("error - returns 404 for non-existent wallet", func(t *testing.T) {
		mockSvc := newMockServiceWithListAndFind(
			func(ctx context.Context, input CreateWalletInput) (*WalletOutput, error) {
				return nil, nil
			},
			nil,
			func(ctx context.Context, id string) (*WalletOutput, error) {
				return nil, errors.New("not found")
			},
		)

		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/wallets/non-existent", nil)
		c.Params = []gin.Param{{Key: "id", Value: "non-existent"}}

		handler.Find(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "WALLET_NOT_FOUND", response.Error.Code)
	})
}

func TestHandler_Update(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - updates wallet", func(t *testing.T) {
		newDesc := "Descrição Atualizada"
		mockSvc := newMockServiceWithUpdate(
			func(ctx context.Context, id string, input UpdateWalletInput) (*WalletOutput, error) {
				return &WalletOutput{
					ID:          id,
					Name:        input.Name,
					Description: input.Description,
					CreatedAt:   "2026-07-25T10:00:00Z",
					UpdatedAt:   "2026-07-25T14:00:00Z",
				}, nil
			},
		)

		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"name":        "Nome Atualizado",
			"description": newDesc,
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/api/v1/wallets/wallet-123", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-123"}}

		handler.Update(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[*WalletOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "wallet-123", response.Data.ID)
		assert.Equal(t, "Nome Atualizado", response.Data.Name)
		assert.Equal(t, "Descrição Atualizada", *response.Data.Description)
	})

	t.Run("error - returns 422 for invalid update input", func(t *testing.T) {
		mockSvc := newMockServiceWithUpdate(nil)
		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"name": "AB",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/api/v1/wallets/wallet-123", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-123"}}

		handler.Update(c)

		assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 404 when wallet not found", func(t *testing.T) {
		mockSvc := newMockServiceWithUpdate(
			func(ctx context.Context, id string, input UpdateWalletInput) (*WalletOutput, error) {
				return nil, errors.New("not found")
			},
		)

		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"name": "Nome",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/api/v1/wallets/wallet-123", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-123"}}

		handler.Update(c)

		assert.Equal(t, http.StatusNotFound, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "WALLET_NOT_FOUND", response.Error.Code)
	})

	t.Run("error - returns 409 when name already exists", func(t *testing.T) {
		mockSvc := newMockServiceWithUpdate(
			func(ctx context.Context, id string, input UpdateWalletInput) (*WalletOutput, error) {
				return nil, ErrWalletNameAlreadyExists
			},
		)

		handler := NewHandler(mockSvc)

		body := map[string]interface{}{
			"name": "Duplicada",
		}
		jsonBody, _ := json.Marshal(body)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("PUT", "/api/v1/wallets/wallet-123", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-123"}}

		handler.Update(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.NotNil(t, response.Error)
		assert.Equal(t, "WALLET_NAME_ALREADY_EXISTS", response.Error.Code)
	})
}
