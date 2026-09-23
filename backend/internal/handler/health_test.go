package handler

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"kanban-dev-app/backend/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func getHealth(t *testing.T, db *sql.DB) (*httptest.ResponseRecorder, healthResponse) {
	t.Helper()
	rec := httptest.NewRecorder()
	NewAPI(db, config.Config{AppEnv: "teste"}).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))

	var got healthResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("decode %q: %v", rec.Body.String(), err)
	}
	return rec, got
}

func TestHealthIsOKWhenTheDatabaseAnswers(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	rec, got := getHealth(t, db)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	if got.Status != "ok" || got.Checks["database"] != "ok" {
		t.Errorf("body = %+v, want status ok and database ok", got)
	}
	if got.Environment != "teste" {
		t.Errorf("environment = %q, want %q", got.Environment, "teste")
	}
}

// O monitor decide pelo status code, então um banco fora do ar não pode
// continuar devolvendo 200 — seria um alerta que nunca dispara.
func TestHealthIs503WhenTheDatabaseIsDown(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	_ = db.Close()

	rec, got := getHealth(t, db)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if got.Status != "degraded" || got.Checks["database"] == "ok" {
		t.Errorf("body = %+v, want status degraded and a failing database check", got)
	}
}

func TestHealthWithoutDatabaseIsDegraded(t *testing.T) {
	rec, got := getHealth(t, nil)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if got.Status != "degraded" {
		t.Errorf("status = %q, want degraded", got.Status)
	}
}

// Uma sondagem a cada minuto com o banco fora geraria um evento por minuto no
// Sentry. Quem avisa é o monitor, não o reporting de erros.
func TestHealthDoesNotReportToSentry(t *testing.T) {
	spy := withSpySentry(t)
	db, err := sql.Open("pgx", "postgres://nobody@127.0.0.1:1/none")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	_ = db.Close()

	if rec, _ := getHealth(t, db); rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rec.Code)
	}
	if n := len(spy.captured()); n != 0 {
		t.Errorf("health check sent %d events to Sentry, want 0", n)
	}
}

// Resposta de monitoramento não pode ser cacheada por proxy nenhum.
func TestHealthIsNotCacheable(t *testing.T) {
	rec, _ := getHealth(t, nil)

	if got := rec.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}
}
