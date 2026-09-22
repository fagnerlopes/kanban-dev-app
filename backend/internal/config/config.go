package config

import (
	"log/slog"
	"net/url"
	"os"
	"strconv"
)

// Config holds runtime configuration read from environment variables.
// Locally, values come from .env; in deployed environments, from Kamal config.
type Config struct {
	// Port the HTTP server listens on. 80 in deployed, 8080 in local dev.
	Port string
	// DatabaseURL is the Postgres connection string.
	DatabaseURL string
	// BaseURL is the public origin of the app (used to derive absolute URLs).
	BaseURL string
	// DevMode enables development-only routes (e.g. POST /api/dev/login).
	// Must NEVER be set in production.
	DevMode bool
	// SentryDSN is the Sentry DSN for error reporting. Empty disables reporting.
	SentryDSN string
	// AppEnv names the environment ("local", "preview", ...). It is what Sentry
	// groups issues by, so it must be a short label -- never a URL.
	AppEnv string
	// SentryTracesSampleRate is the fraction of requests traced, 0 to 1.
	// 0 disables tracing. The workshop default is 1 (trace everything) because
	// the traffic is a handful of people in a room; a real product would sample.
	SentryTracesSampleRate float64
}

// Load reads configuration from the environment.
func Load() Config {
	cfg := Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: normalizeDatabaseURL(os.Getenv("DATABASE_URL")),
		BaseURL:     getenv("BASE_URL", "http://localhost:5173"),
		DevMode:     os.Getenv("DEV_MODE") == "1",
		SentryDSN:   os.Getenv("SENTRY_DSN"),
		AppEnv:      getenv("APP_ENV", "local"),

		SentryTracesSampleRate: sampleRate(os.Getenv("SENTRY_TRACES_SAMPLE_RATE"), 1),
	}
	return cfg
}

// sampleRate parses a 0..1 rate, falling back to the default for anything
// unparseable or out of range -- a typo here should not silently turn tracing
// off (or, worse, be rejected and take the whole app down at startup).
func sampleRate(raw string, fallback float64) float64 {
	if raw == "" {
		return fallback
	}
	v, err := strconv.ParseFloat(raw, 64)
	if err != nil || v < 0 || v > 1 {
		slog.Warn("invalid SENTRY_TRACES_SAMPLE_RATE, using default",
			"value", raw, "default", fallback)
		return fallback
	}
	return v
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// normalizeDatabaseURL re-serializes the userinfo (user:password) of a
// postgres:// DSN so the password is always percent-encoded. This makes the
// DSN safe for net/url / pgx even if the password contains characters like
// '@', '/', '+' or '#'. If the URL cannot be parsed (e.g. the password
// already broke parsing), the input is returned unchanged so the driver
// surfaces the underlying error.
func normalizeDatabaseURL(dsn string) string {
	if dsn == "" {
		return dsn
	}
	u, err := url.Parse(dsn)
	if err != nil || u.User == nil {
		return dsn
	}
	pass, _ := u.User.Password()
	u.User = url.UserPassword(u.User.Username(), pass)
	return u.String()
}
