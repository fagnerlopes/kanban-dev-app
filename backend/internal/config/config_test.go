package config

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	for _, k := range []string{"PORT", "BASE_URL", "DEV_MODE", "SENTRY_DSN", "DATABASE_URL"} {
		t.Setenv(k, "")
	}

	cfg := Load()
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want the local dev default 8080", cfg.Port)
	}
	if cfg.BaseURL != "http://localhost:5173" {
		t.Errorf("BaseURL = %q, want the Vite dev origin", cfg.BaseURL)
	}
	if cfg.DevMode {
		t.Error("DevMode = true with DEV_MODE unset; it must default to off")
	}
}

func TestDevModeOnlyOnExplicitOne(t *testing.T) {
	for value, want := range map[string]bool{"1": true, "": false, "0": false, "true": false} {
		t.Setenv("DEV_MODE", value)
		if got := Load().DevMode; got != want {
			t.Errorf("DEV_MODE=%q -> DevMode = %v, want %v", value, got, want)
		}
	}
}

// A generated Postgres password can contain "@", which pgx would otherwise
// read as the start of the host.
func TestNormalizeDatabaseURLEncodesPassword(t *testing.T) {
	got := normalizeDatabaseURL("postgres://postgres:p@ss@db:5432/postgres?sslmode=disable")

	if strings.Contains(got, "p@ss@") {
		t.Fatalf("password left unencoded in %q", got)
	}
	u, err := url.Parse(got)
	if err != nil {
		t.Fatalf("normalized DSN no longer parses: %v", err)
	}
	if pass, _ := u.User.Password(); pass != "p@ss" {
		t.Errorf("password round-tripped as %q, want %q", pass, "p@ss")
	}
	if u.Host != "db:5432" || u.Path != "/postgres" {
		t.Errorf("host/database changed: host=%q path=%q", u.Host, u.Path)
	}
	if u.Query().Get("sslmode") != "disable" {
		t.Errorf("query parameters lost: %q", u.RawQuery)
	}
}

// When the DSN cannot be parsed, it must pass through untouched so pgx reports
// the real error instead of a mangled one.
func TestNormalizeDatabaseURLPassesThroughUnparseable(t *testing.T) {
	for _, dsn := range []string{"", "not a url at all", "postgres://postgres:pa/ss@db:5432/postgres"} {
		if got := normalizeDatabaseURL(dsn); got != dsn {
			t.Errorf("normalizeDatabaseURL(%q) = %q, want it returned unchanged", dsn, got)
		}
	}
}

// DEV_MODE is a local-only flag: it registers POST /api/dev/login and tells the
// server that Vite serves the SPA. A deploy config that sets it takes the whole
// app down — every page 404s while /up and the API still look healthy. This
// guards the exact regression.
func TestDeployConfigsNeverSetDevMode(t *testing.T) {
	configs, err := filepath.Glob(filepath.Join("..", "..", "..", "config", "deploy*.yml"))
	if err != nil {
		t.Fatalf("glob deploy configs: %v", err)
	}
	if len(configs) == 0 {
		t.Fatal("no config/deploy*.yml found — the guard would silently pass")
	}

	for _, path := range configs {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		for i, line := range strings.Split(string(body), "\n") {
			code, _, _ := strings.Cut(line, "#") // ignore comments
			if strings.Contains(code, "DEV_MODE") {
				t.Errorf("%s:%d sets DEV_MODE — it must never be set in a deployed environment: %s",
					path, i+1, strings.TrimSpace(line))
			}
		}
	}
}
