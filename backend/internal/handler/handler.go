package handler

import (
	"database/sql"
	"log/slog"
	"net/http"

	"kanban-dev-app/backend/internal/config"

	"github.com/getsentry/sentry-go"
)

// API bundles the dependencies shared by all handlers.
type API struct {
	db  *sql.DB
	cfg config.Config
}

// NewAPI constructs the API handler with its dependencies.
func NewAPI(db *sql.DB, cfg config.Config) http.Handler {
	api := &API{db: db, cfg: cfg}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/board", api.handleBoard)
	mux.HandleFunc("POST /api/tasks", api.handleCreateTask)
	mux.HandleFunc("PATCH /api/tasks/{id}", api.handleUpdateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", api.handleDeleteTask)
	if cfg.DevMode {
		mux.HandleFunc("POST /api/dev/login", api.handleDevLogin)
	}
	return recoverWithSentry(mux)
}

// InitSentry initializes the Sentry SDK. It is a no-op (with a log) when the
// DSN is empty, so the app runs fine without Sentry configured.
func InitSentry(dsn, envName string) error {
	if dsn == "" {
		return nil
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         dsn,
		Environment: envName,
		Release:     "kanban-dev-app",
	})
	if err != nil {
		return err
	}
	slog.Info("sentry ready", "environment", envName)
	return nil
}

// recoverWithSentry wraps a handler, capturing any panic to Sentry and
// returning a 500. This is the hook the planted backend bug will trip.
func recoverWithSentry(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if p := recover(); p != nil {
				sentry.CaptureException(panicToError(p))
				slog.Error("panic recovered", "path", r.URL.Path, "panic", p)
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
