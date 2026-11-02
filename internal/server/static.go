package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var staticFS embed.FS

// RegisterStaticRoutes mounts the embedded React 19 single-page application.
func (s *Server) RegisterStaticRoutes() {
	distSub, err := fs.Sub(staticFS, "dist")
	if err != nil {
		return
	}

	fileServer := http.FileServer(http.FS(distSub))

	s.mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		// Do not intercept API requests
		if strings.HasPrefix(r.URL.Path, "/api") {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Check if file exists in embedded filesystem
		f, err := distSub.Open(path)
		if err == nil {
			_ = f.Close()
			fileServer.ServeHTTP(w, r)
			return
		}

		// Fallback to index.html for client-side SPA routing
		indexBytes, err := fs.ReadFile(distSub, "index.html")
		if err == nil {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(indexBytes)
			return
		}

		http.NotFound(w, r)
	})
}
