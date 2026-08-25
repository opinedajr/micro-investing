package stock

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/opinedajr/micro-investing/internal/shared/api"
	"github.com/stretchr/testify/assert"
)

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - registers stock routes under /api/v1", func(t *testing.T) {
		service := newMockServiceWithList(func(ctx context.Context) ([]StockOutput, error) {
			return []StockOutput{
				{ID: "1", Ticker: "PETR4", Name: "Petrobras PN", Sector: "Petróleo", Rank: 10},
			}, nil
		})

		handler := NewHandler(service)
		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/v1/stocks", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response api.Response[[]StockOutput]
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Len(t, response.Data, 1)
		assert.Equal(t, "PETR4", response.Data[0].Ticker)
	})
}
