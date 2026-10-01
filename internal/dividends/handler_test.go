package dividends

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
	"github.com/stretchr/testify/require"
)

func TestHandler_Create(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - creates dividend", func(t *testing.T) {
		mockSvc := newMockService(serviceFuncs{
			createFunc: func(ctx context.Context, input CreateDividendInput) (*DividendOutput, error) {
				return &DividendOutput{ID: "dividend-id", Year: input.Year, Amount: input.Amount}, nil
			},
		})
		handler := NewHandler(mockSvc)

		jsonBody, _ := json.Marshal(map[string]interface{}{
			"year":   2026,
			"amount": 150000,
		})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/dividends", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response api.Response[*DividendOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "dividend-id", response.Data.ID)
		assert.Equal(t, 2026, response.Data.Year)
		assert.Equal(t, int64(150000), response.Data.Amount)
	})

	t.Run("error - returns 400 for invalid json", func(t *testing.T) {
		handler := NewHandler(newMockService(serviceFuncs{}))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/dividends", bytes.NewBuffer([]byte("{invalid")))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 400 for missing fields", func(t *testing.T) {
		handler := NewHandler(newMockService(serviceFuncs{}))

		jsonBody, _ := json.Marshal(map[string]interface{}{"year": 2026})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/dividends", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "VALIDATION_ERROR", response.Error.Code)
	})

	t.Run("error - returns 400 for invalid year from service", func(t *testing.T) {
		mockSvc := newMockService(serviceFuncs{
			createFunc: func(ctx context.Context, input CreateDividendInput) (*DividendOutput, error) {
				return nil, ErrInvalidDividendYear
			},
		})
		handler := NewHandler(mockSvc)

		jsonBody, _ := json.Marshal(map[string]interface{}{"year": 1500, "amount": 100})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/dividends", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
	})

	t.Run("error - returns 409 when dividend already exists", func(t *testing.T) {
		mockSvc := newMockService(serviceFuncs{
			createFunc: func(ctx context.Context, input CreateDividendInput) (*DividendOutput, error) {
				return nil, ErrDividendAlreadyExists
			},
		})
		handler := NewHandler(mockSvc)

		jsonBody, _ := json.Marshal(map[string]interface{}{"year": 2026, "amount": 150000})

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("POST", "/api/v1/wallets/wallet-id/dividends", bytes.NewBuffer(jsonBody))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.Create(c)

		assert.Equal(t, http.StatusConflict, w.Code)

		var response api.Response[interface{}]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, "DIVIDEND_ALREADY_EXISTS", response.Error.Code)
	})
}

func TestHandler_List(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - returns items", func(t *testing.T) {
		mockSvc := newMockService(serviceFuncs{
			listFunc: func(ctx context.Context, filter DividendFilter) ([]DividendOutput, error) {
				return []DividendOutput{
					{ID: "dividend-1", Year: 2026, Amount: 150000},
				}, nil
			},
		})
		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dividends", nil)
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.List(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[*DividendsListOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		require.Len(t, response.Data.Items, 1)
		assert.Equal(t, 2026, response.Data.Items[0].Year)
	})

	t.Run("success - returns empty items for filter without match", func(t *testing.T) {
		mockSvc := newMockService(serviceFuncs{
			listFunc: func(ctx context.Context, filter DividendFilter) ([]DividendOutput, error) {
				return []DividendOutput{}, nil
			},
		})
		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dividends?year=1990", nil)
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.List(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[*DividendsListOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		require.Empty(t, response.Data.Items)
	})

	t.Run("error - returns 500 on service failure", func(t *testing.T) {
		mockSvc := newMockService(serviceFuncs{
			listFunc: func(ctx context.Context, filter DividendFilter) ([]DividendOutput, error) {
				return nil, assert.AnError
			},
		})
		handler := NewHandler(mockSvc)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/api/v1/wallets/wallet-id/dividends", nil)
		c.Params = []gin.Param{{Key: "id", Value: "wallet-id"}}

		handler.List(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}
