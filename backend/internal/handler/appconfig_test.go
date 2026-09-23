package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"kanban-dev-app/backend/internal/config"
)

func serveConfig(t *testing.T, cfg config.Config) (*httptest.ResponseRecorder, appConfig) {
	t.Helper()
	rec := httptest.NewRecorder()
	NewAPI(nil, cfg).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/config", nil))

	var got appConfig
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode body %q: %v", rec.Body.String(), err)
	}
	return rec, got
}

func TestConfigExposesSentryDSN(t *testing.T) {
	const dsn = "https://publickey@o123.ingest.sentry.io/456"

	rec, got := serveConfig(t, config.Config{SentryDSN: dsn, AppEnv: "preview"})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got.SentryDSN != dsn {
		t.Errorf("sentry_dsn = %q, want %q", got.SentryDSN, dsn)
	}
	if got.Environment != "preview" {
		t.Errorf("environment = %q, want %q", got.Environment, "preview")
	}
	if got.Release == "" {
		t.Error("release is empty; Sentry needs it to line up backend and browser events")
	}
}

func TestConfigWithoutDSNStillAnswers(t *testing.T) {
	rec, got := serveConfig(t, config.Config{AppEnv: "local"})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200 even without a DSN", rec.Code)
	}
	if got.SentryDSN != "" {
		t.Errorf("sentry_dsn = %q, want empty", got.SentryDSN)
	}
}

func TestConfigLeaksNoSecrets(t *testing.T) {
	cfg := config.Config{
		SentryDSN:   "https://publickey@o123.ingest.sentry.io/456",
		AppEnv:      "preview",
		DatabaseURL: "postgres://postgres:sup3r-s3cret@db:5432/postgres?sslmode=disable",
	}

	rec, _ := serveConfig(t, cfg)

	if body := rec.Body.String(); strings.Contains(body, "sup3r-s3cret") {
		t.Errorf("/api/config leaked the database password: %s", body)
	}
}
