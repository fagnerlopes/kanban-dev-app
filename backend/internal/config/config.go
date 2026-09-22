package config

import (
	"net/url"
	"os"
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
	}
	return cfg
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
