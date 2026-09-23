package handler

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// RegisterFrontend serve a SPA construída. Devolve false quando distDir
// não existe.
func RegisterFrontend(mux *http.ServeMux, distDir string) bool {
	if _, err := os.Stat(distDir); err != nil {
		slog.Warn("frontend dist not found — SPA routes will return 404", "path", distDir)
		return false
	}

	spaShell := "index.html"
	if _, err := os.Stat(filepath.Join(distDir, "__spa-fallback.html")); err == nil {
		spaShell = "__spa-fallback.html"
	}

	fs := http.FileServer(http.Dir(distDir))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/auth/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/" {
			fs.ServeHTTP(w, r)
			return
		}
		target := filepath.Join(distDir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(target); err == nil {
			if !info.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}
			if page := filepath.Join(target, "index.html"); fileExists(page) {
				http.ServeFile(w, r, page)
				return
			}
			// Directory without index.html (e.g. /assets/): fall through to the shell.
		}
		http.ServeFile(w, r, filepath.Join(distDir, spaShell))
	})
	return true
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
