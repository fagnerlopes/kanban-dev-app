package handler

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// RegisterFrontend serves the built SPA from distDir on the given mux.
//
// distDir must be the literal "frontend/dist": the binary's working directory
// in production is the Dockerfile WORKDIR, not backend/.
//
// Three rules are encoded here, each learned the hard way:
//   - real files go through http.FileServer, so assets get the correct
//     Content-Type (serving the SPA shell for every path makes browsers reject
//     .js/.css with a MIME type error);
//   - prerendered pages ("about/index.html") are served directly, because
//     handing the directory to FileServer 301-redirects every URL to its
//     trailing-slash form;
//   - directories are never handed to FileServer, which would expose a
//     browsable listing of the app's internals.
//
// Returns false (and logs) when distDir does not exist, so the caller can tell
// a missing build from a successful registration.
func RegisterFrontend(mux *http.ServeMux, distDir string) bool {
	if _, err := os.Stat(distDir); err != nil {
		slog.Warn("frontend dist not found — SPA routes will return 404", "path", distDir)
		return false
	}

	// SPA shell: __spa-fallback.html when the build prerenders "/", else index.html.
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
		// "/" always maps to index.html (the prerendered home in prerendered builds).
		if r.URL.Path == "/" {
			fs.ServeHTTP(w, r)
			return
		}
		target := filepath.Join(distDir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(target); err == nil {
			if !info.IsDir() {
				fs.ServeHTTP(w, r) // real asset; FileServer sets Content-Type
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
