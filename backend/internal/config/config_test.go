package config

import (
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	for _, k := range []string{"PORT", "BASE_URL", "DEV_MODE", "SENTRY_DSN", "DATABASE_URL", "APP_ENV"} {
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
	if cfg.AppEnv != "local" {
		t.Errorf("AppEnv = %q, want the local default", cfg.AppEnv)
	}
}

// A typo in the sample rate must not disable tracing silently, and must never
// take the app down at startup — it is an observability knob, not a dependency.
func TestSentryTracesSampleRateFallsBackOnBadInput(t *testing.T) {
	for _, raw := range []string{"", "abc", "-0.5", "2", "1,0"} {
		t.Setenv("SENTRY_TRACES_SAMPLE_RATE", raw)
		if got := Load().SentryTracesSampleRate; got != 1 {
			t.Errorf("SENTRY_TRACES_SAMPLE_RATE=%q -> %v, want the default 1", raw, got)
		}
	}
}

func TestSentryTracesSampleRateAcceptsValidRates(t *testing.T) {
	for raw, want := range map[string]float64{"0": 0, "0.25": 0.25, "1": 1} {
		t.Setenv("SENTRY_TRACES_SAMPLE_RATE", raw)
		if got := Load().SentryTracesSampleRate; got != want {
			t.Errorf("SENTRY_TRACES_SAMPLE_RATE=%q -> %v, want %v", raw, got, want)
		}
	}
}

// AppEnv is what Sentry groups issues by. It used to receive BASE_URL, which
// turned the environment facet into a URL and made the Sentry UI useless.
func TestAppEnvComesFromAppEnvNotBaseURL(t *testing.T) {
	t.Setenv("APP_ENV", "preview")
	t.Setenv("BASE_URL", "https://191.252.226.176.nip.io")

	if got := Load().AppEnv; got != "preview" {
		t.Errorf("AppEnv = %q, want %q", got, "preview")
	}
}

// Every deployed environment must label itself, otherwise Sentry files preview
// issues under "local" and the workshop participant cannot tell them apart.
func TestDeployConfigsSetAppEnv(t *testing.T) {
	forEachEnvConfig(t, func(t *testing.T, path, body string) {
		if !strings.Contains(body, "APP_ENV:") {
			t.Errorf("%s does not set APP_ENV — Sentry would label its issues %q", path, "local")
		}
	})
}

// The public hostname must stay driven by the APP_DOMAIN repository variable.
// Hardcoding a domain works for whoever typed it and breaks every fork, which
// would route a name its owner does not control.
func TestDeployConfigsKeepDomainConfigurable(t *testing.T) {
	forEachEnvConfig(t, func(t *testing.T, path, body string) {
		if !strings.Contains(body, "APP_DOMAIN") {
			t.Errorf("%s does not read APP_DOMAIN — a fork could not set its own domain", path)
		}
	})
}

// The nip.io host must be routed unconditionally, not only when APP_DOMAIN is
// empty. kamal-proxy routes strictly by Host header, so a config that swapped
// nip.io for the custom domain takes the app offline the moment the domain is
// set before its DNS is ready — and the deploy still reports success, because
// the health check talks to the container directly and never exercises the
// public hostname. Verified the hard way: a green deploy left the app
// unreachable on every address.
func TestDeployConfigsAlwaysRouteNipIo(t *testing.T) {
	forEachEnvConfig(t, func(t *testing.T, path, body string) {
		if !strings.Contains(body, "nip.io") {
			t.Fatalf("%s never mentions nip.io — the IP-address URL would stop working", path)
		}
		// The nip.io host must not sit behind an "only if APP_DOMAIN is empty"
		// branch. Such a branch reads as an assignment guarded by `if`/`unless`
		// on the same line as the nip.io literal.
		for i, line := range strings.Split(body, "\n") {
			if !strings.Contains(line, "nip.io") {
				continue
			}
			if strings.Contains(line, " if ") || strings.Contains(line, " unless ") {
				t.Errorf("%s:%d makes the nip.io host conditional: %s",
					path, i+1, strings.TrimSpace(line))
			}
		}
	})
}

func forEachEnvConfig(t *testing.T, check func(t *testing.T, path, body string)) {
	t.Helper()
	configs, err := filepath.Glob(filepath.Join("..", "..", "..", "config", "deploy.*.yml"))
	if err != nil {
		t.Fatalf("glob deploy configs: %v", err)
	}
	if len(configs) == 0 {
		t.Fatal("no config/deploy.<env>.yml found — the guard would silently pass")
	}
	for _, path := range configs {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		check(t, path, string(body))
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
