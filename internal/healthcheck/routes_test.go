package healthcheck

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success - registers health route under /api/v1", func(t *testing.T) {
		mockService := &MockService{}
		handler := NewHandler(mockService)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response []Health
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if len(response) == 0 {
			t.Fatal("expected at least 1 health result, got 0")
		}

		if response[0].ServiceName != ServiceName {
			t.Errorf("expected service name %s, got %s", ServiceName, response[0].ServiceName)
		}
	})

	t.Run("success - returns 404 on unmounted root path", func(t *testing.T) {
		mockService := &MockService{}
		handler := NewHandler(mockService)

		r := gin.New()
		v1 := r.Group("/api/v1")
		RegisterRoutes(v1, handler)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404 for /health (route moved to /api/v1/health), got %d", w.Code)
		}
	})
}
