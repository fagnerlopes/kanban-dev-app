package handler

import (
	"database/sql"
	"log/slog"
	"net/http"

	"kanban-dev-app/backend/internal/config"

	"github.com/getsentry/sentry-go"
)

// releaseName tags every Sentry event (backend and browser) with the same
// release, so both sides of an issue line up in the Sentry UI.
const releaseName = "kanban-dev-app"

// API bundles the dependencies shared by all handlers.
type API struct {
	db  *sql.DB
	cfg config.Config
}

// NewAPI constructs the API handler with its dependencies.
func NewAPI(db *sql.DB, cfg config.Config) http.Handler {
	api := &API{db: db, cfg: cfg}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/config", api.handleConfig)
	mux.HandleFunc("GET /api/board", api.handleBoard)
	mux.HandleFunc("POST /api/tasks", api.handleCreateTask)
	mux.HandleFunc("PATCH /api/tasks/{id}", api.handleUpdateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", api.handleDeleteTask)
	if cfg.DevMode {
		mux.HandleFunc("POST /api/dev/login", api.handleDevLogin)
	}
	return recoverWithSentry(mux)
}

// InitSentry initializes the Sentry SDK. It is a no-op when the DSN is empty,
// so the app runs fine without Sentry configured.
//
// Note on the options: Sentry's onboarding snippet for Go suggests
// `EnableLogs: true`, which does not compile against this SDK. That flag became
// `DisableLogs` and was then removed entirely (see the SDK changelog) -- logs
// and metrics are now switched on simply by *using* their APIs
// (sentry.NewLogger / sentry.NewMeter), which is what this app does. Tracing is
// the one that still needs a flag, plus a sample rate.
func InitSentry(cfg config.Config) error {
	if cfg.SentryDSN == "" {
		return nil
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.SentryDSN,
		Environment: cfg.AppEnv,
		Release:     releaseName,
		// Stack traces on captured messages, not just on exceptions --
		// otherwise a reported 500 arrives with nowhere to look.
		AttachStacktrace: true,
		EnableTracing:    cfg.SentryTracesSampleRate > 0,
		TracesSampleRate: cfg.SentryTracesSampleRate,
	})
	if err != nil {
		return err
	}
	slog.Info("sentry ready",
		"environment", cfg.AppEnv,
		"traces_sample_rate", cfg.SentryTracesSampleRate)
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
