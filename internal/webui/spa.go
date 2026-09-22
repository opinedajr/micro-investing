package webui

import (
	"io/fs"
	"net/http"
	"strings"
)

type SPAHandler struct {
	root http.FileSystem
}

func NewSPAHandler(fsys fs.FS) *SPAHandler {
	return &SPAHandler{root: http.FS(fsys)}
}

func (s *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path

	if strings.HasPrefix(path, "/api/") {
		http.NotFound(w, r)
		return
	}

	if path != "/" {
		if f, err := s.root.Open(path); err == nil {
			f.Close()
			if strings.HasPrefix(path, "/assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			http.FileServer(s.root).ServeHTTP(w, r)
			return
		}
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	r.URL.Path = "/"
	http.FileServer(s.root).ServeHTTP(w, r)
}
