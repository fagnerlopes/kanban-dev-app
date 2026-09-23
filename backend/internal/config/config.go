package config

import (
	"log/slog"
	"net/url"
	"os"
	"strconv"
)

// Config holds runtime configuration read from environment variables.
type Config struct {
	Port        string
	DatabaseURL string
	BaseURL     string
	// DevMode nunca pode ser ligado em ambiente publicado.
	DevMode   bool
	SentryDSN string
	// AppEnv é o rótulo curto de ambiente que o Sentry usa para agrupar issues.
	AppEnv string
	// SentryTracesSampleRate vai de 0 a 1; 0 desliga o tracing.
	SentryTracesSampleRate float64
}

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

// sampleRate cai no padrão para qualquer valor inválido: um erro de digitação
// aqui não pode derrubar o app nem desligar o tracing em silêncio.
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

// normalizeDatabaseURL garante que a senha fique percent-encoded, para que
// caracteres como '@' não quebrem o parsing da DSN.
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
