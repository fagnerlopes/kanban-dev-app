package config

import "os"

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
}

// Load reads configuration from the environment.
func Load() Config {
	cfg := Config{
		Port:        getenv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		BaseURL:     getenv("BASE_URL", "http://localhost:5173"),
		DevMode:     os.Getenv("DEV_MODE") == "1",
		SentryDSN:   os.Getenv("SENTRY_DSN"),
	}
	return cfg
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
