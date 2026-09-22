package webui

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
)

func newDistFS() fstest.MapFS {
	return fstest.MapFS{
		"index.html":           {Data: []byte("<!doctype html><html><body><div id=\"app\"></div></body></html>")},
		"favicon.svg":          {Data: []byte("<svg></svg>")},
		"assets/app-9f8e7d.js": {Data: []byte("console.log('app')")},
	}
}

func performSPARequest(handler *SPAHandler, method string, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, path, nil)
	handler.ServeHTTP(w, req)
	return w
}

func TestSPAHandler(t *testing.T) {
	handler := NewSPAHandler(newDistFS())

	t.Run("success - serves index html at root with no cache headers", func(t *testing.T) {
		w := performSPARequest(handler, http.MethodGet, "/")

		if w.Code != http.StatusOK {
			t.Fatalf("expected status code 200, got %d", w.Code)
		}
		if contentType := w.Header().Get("Content-Type"); !strings.Contains(contentType, "text/html") {
			t.Errorf("expected content-type text/html, got %s", contentType)
		}
		if !strings.Contains(w.Body.String(), `<div id="app">`) {
			t.Errorf("expected index html with app root div, got %s", w.Body.String())
		}
		if cache := w.Header().Get("Cache-Control"); cache != "no-cache, no-store, must-revalidate" {
			t.Errorf("expected no-cache cache-control on index, got %q", cache)
		}
	})

	t.Run("success - falls back to index html for unknown spa routes", func(t *testing.T) {
		w := performSPARequest(handler, http.MethodGet, "/some/unknown/route")

		if w.Code != http.StatusOK {
			t.Fatalf("expected status code 200, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), `<div id="app">`) {
			t.Errorf("expected spa fallback to index html, got %s", w.Body.String())
		}
		if cache := w.Header().Get("Cache-Control"); cache != "no-cache, no-store, must-revalidate" {
			t.Errorf("expected no-cache cache-control on fallback, got %q", cache)
		}
	})

	t.Run("success - serves static assets", func(t *testing.T) {
		w := performSPARequest(handler, http.MethodGet, "/favicon.svg")

		if w.Code != http.StatusOK {
			t.Fatalf("expected status code 200, got %d", w.Code)
		}
	})

	t.Run("success - serves hashed assets with immutable cache header", func(t *testing.T) {
		w := performSPARequest(handler, http.MethodGet, "/assets/app-9f8e7d.js")

		if w.Code != http.StatusOK {
			t.Fatalf("expected status code 200, got %d", w.Code)
		}
		if cache := w.Header().Get("Cache-Control"); cache != "public, max-age=31536000, immutable" {
			t.Errorf("expected immutable cache-control on hashed asset, got %q", cache)
		}
	})

	t.Run("error - returns not found for api paths", func(t *testing.T) {
		w := performSPARequest(handler, http.MethodGet, "/api/v1/unknown")

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected status code 404, got %d", w.Code)
		}
	})
}
