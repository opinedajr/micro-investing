package webui

import (
	"encoding/json"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	r := gin.New()
	RegisterRoutes(r)
	return r
}

func TestRegisterRoutes(t *testing.T) {
	t.Run("success - embedded dist is resolvable", func(t *testing.T) {
		dist, err := fs.Sub(DistFS, "dist")
		if err != nil {
			t.Fatalf("expected embedded dist to resolve, got error: %v", err)
		}
		if _, err := dist.Open("."); err != nil {
			t.Fatalf("expected embedded dist to be readable, got error: %v", err)
		}
	})

	t.Run("error - returns json not found for unknown api routes", func(t *testing.T) {
		r := setupRouter(t)

		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil)
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status code 404, got %d", w.Code)
		}
		if contentType := w.Header().Get("Content-Type"); !strings.Contains(contentType, "application/json") {
			t.Errorf("expected content-type application/json, got %s", contentType)
		}

		var response struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if response.Error.Code != "NOT_FOUND" {
			t.Errorf("expected error code NOT_FOUND, got %s", response.Error.Code)
		}
	})
}
