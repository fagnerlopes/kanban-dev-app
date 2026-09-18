package handler

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"kanban-dev-app/backend/internal/config"
	"kanban-dev-app/backend/internal/database"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// testEnv bundles a database and an HTTP handler for integration tests.
type testEnv struct {
	db   *sql.DB
	root http.Handler
}

// setupTest connects to the test database (DATABASE_URL), runs migrations,
// and builds the API handler. It returns a cleanup func that restores the
// tables to an empty state so each test starts from a clean slate.
func setupTest(t *testing.T) *testEnv {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set — skipping DB integration test")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.Ping(); err != nil {
		t.Fatalf("ping db: %v", err)
	}

	if err := database.Migrate(context.Background(), db); err != nil {
		t.Fatalf("migrations: %v", err)
	}

	// Reset tables so every test starts clean. Delete tasks first (FK child),
	// then columns, then re-seed the default columns.
	if _, err := db.Exec("DELETE FROM tasks"); err != nil {
		t.Fatalf("delete tasks: %v", err)
	}
	if _, err := db.Exec("DELETE FROM columns"); err != nil {
		t.Fatalf("delete columns: %v", err)
	}
	if _, err := db.Exec(`
		INSERT INTO columns (name, position) VALUES
			('Backlog', 0), ('To Do', 1), ('In Dev', 2), ('Review', 3), ('Done', 4)
		ON CONFLICT (name) DO NOTHING`); err != nil {
		t.Fatalf("seed columns: %v", err)
	}

	cfg := config.Config{DevMode: true}
	return &testEnv{db: db, root: NewAPI(db, cfg)}
}

// do performs an HTTP request against the API and returns the response.
func (e *testEnv) do(t *testing.T, method, target string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		rdr = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, target, rdr)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	e.root.ServeHTTP(rec, req)
	return rec
}
