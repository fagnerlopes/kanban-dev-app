package handler

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// buildDist writes a minimal SPA build tree mirroring what `react-router build`
// emits into frontend/build/client.
func buildDist(t *testing.T, shellName string) string {
	t.Helper()
	dir := t.TempDir()
	write := func(rel, body string) {
		t.Helper()
		full := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", rel, err)
		}
		if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
			t.Fatalf("write %s: %v", rel, err)
		}
	}
	write("index.html", "<html>home</html>")
	if shellName != "index.html" {
		write(shellName, "<html>shell</html>")
	}
	write("assets/app-abc123.js", "console.log('app')")
	write("assets/app-abc123.css", "body{}")
	write("about/index.html", "<html>about</html>")
	return dir
}

func serve(t *testing.T, dir, path string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	if !RegisterFrontend(mux, dir) {
		t.Fatalf("RegisterFrontend reported a missing dist at %s", dir)
	}
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

// The regression that took the deployed app down: with the SPA handler not
// registered, "/" returned 404 even though the API and /up were healthy.
func TestFrontendServesIndexAtRoot(t *testing.T) {
	dir := buildDist(t, "index.html")
	rec := serve(t, dir, "/")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET / = %d, want 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "home") {
		t.Fatalf("GET / body = %q, want the home document", rec.Body.String())
	}
}

// Serving the SPA shell for hashed assets makes browsers reject them with a
// MIME type error — the app looks completely broken in production only.
func TestFrontendServesAssetsWithCorrectContentType(t *testing.T) {
	dir := buildDist(t, "index.html")
	for path, wantType := range map[string]string{
		"/assets/app-abc123.js":  "javascript",
		"/assets/app-abc123.css": "css",
	} {
		rec := serve(t, dir, path)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s = %d, want 200", path, rec.Code)
			continue
		}
		if got := rec.Header().Get("Content-Type"); !strings.Contains(got, wantType) {
			t.Errorf("GET %s Content-Type = %q, want it to contain %q", path, got, wantType)
		}
	}
}

func TestFrontendFallsBackToShellForClientRoutes(t *testing.T) {
	dir := buildDist(t, "index.html")
	rec := serve(t, dir, "/board")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /board = %d, want 200 (SPA shell)", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "home") {
		t.Fatalf("GET /board body = %q, want the SPA shell", rec.Body.String())
	}
}

// A prerendered page must be served directly: delegating the directory to
// FileServer would 301 to the trailing-slash form.
func TestFrontendServesPrerenderedPageWithoutRedirect(t *testing.T) {
	dir := buildDist(t, "index.html")
	rec := serve(t, dir, "/about")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /about = %d, want 200 (no redirect)", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "about") {
		t.Fatalf("GET /about body = %q, want the prerendered page", rec.Body.String())
	}
}

// Builds that prerender "/" emit the shell as __spa-fallback.html; "/" must
// still serve index.html (the prerendered home), not the shell.
func TestFrontendPrefersSpaFallbackShell(t *testing.T) {
	dir := buildDist(t, "__spa-fallback.html")
	if body := serve(t, dir, "/").Body.String(); !strings.Contains(body, "home") {
		t.Errorf("GET / body = %q, want the prerendered home", body)
	}
	if body := serve(t, dir, "/board").Body.String(); !strings.Contains(body, "shell") {
		t.Errorf("GET /board body = %q, want the SPA fallback shell", body)
	}
}

// A directory without index.html must not render a browsable listing of the
// app's internals.
func TestFrontendNeverListsDirectories(t *testing.T) {
	dir := buildDist(t, "index.html")
	body := serve(t, dir, "/assets").Body.String()
	if strings.Contains(body, "app-abc123.js") {
		t.Fatalf("GET /assets listed the directory contents: %q", body)
	}
}

// Unmatched API paths must 404, not fall back to the SPA shell.
func TestFrontendDoesNotSwallowAPIPaths(t *testing.T) {
	dir := buildDist(t, "index.html")
	for _, path := range []string{"/api/missing", "/auth/callback"} {
		if code := serve(t, dir, path).Code; code != http.StatusNotFound {
			t.Errorf("GET %s = %d, want 404", path, code)
		}
	}
}

func TestFrontendReportsMissingDist(t *testing.T) {
	if RegisterFrontend(http.NewServeMux(), filepath.Join(t.TempDir(), "nope")) {
		t.Fatal("RegisterFrontend reported success for a missing dist directory")
	}
}
