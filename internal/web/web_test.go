package web

import (
	"encoding/json"
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

func performRequest(r *gin.Engine, method string, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	r.ServeHTTP(w, req)
	return w
}

func TestRegisterRoutes(t *testing.T) {
	t.Run("success - serves index html at root", func(t *testing.T) {
		r := setupRouter(t)

		w := performRequest(r, http.MethodGet, "/")

		if w.Code != http.StatusOK {
			t.Fatalf("expected status code 200, got %d", w.Code)
		}
		if contentType := w.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
			t.Errorf("expected content-type text/html, got %s", contentType)
		}
		if !strings.Contains(w.Body.String(), `<div id="app">`) {
			t.Errorf("expected index html with app root div, got %s", w.Body.String())
		}
	})

	t.Run("success - falls back to index html for unknown spa routes", func(t *testing.T) {
		r := setupRouter(t)

		w := performRequest(r, http.MethodGet, "/some/unknown/route")

		if w.Code != http.StatusOK {
			t.Fatalf("expected status code 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), `<div id="app">`) {
			t.Errorf("expected spa fallback to index html, got %s", w.Body.String())
		}
	})

	t.Run("success - serves static assets", func(t *testing.T) {
		r := setupRouter(t)

		w := performRequest(r, http.MethodGet, "/favicon.svg")

		if w.Code != http.StatusOK {
			t.Fatalf("expected status code 200, got %d", w.Code)
		}
	})

	t.Run("error - returns json not found for unknown api routes", func(t *testing.T) {
		r := setupRouter(t)

		w := performRequest(r, http.MethodGet, "/api/v1/unknown")

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
