package handler

import (
	"database/sql"
	"log/slog"
	"net/http"
	"time"

	"kanban-dev-app/backend/internal/config"

	"github.com/getsentry/sentry-go"
)

const releaseName = "kanban-dev-app"

// API bundles the dependencies shared by all handlers.
type API struct {
	db        *sql.DB
	cfg       config.Config
	startedAt time.Time
}

func NewAPI(db *sql.DB, cfg config.Config) http.Handler {
	api := &API{db: db, cfg: cfg, startedAt: time.Now()}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", api.handleHealth)
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

// InitSentry é um no-op quando o DSN está vazio.
//
// Não existe opção EnableLogs neste SDK: logs e métricas ligam pelo uso de
// sentry.NewLogger e sentry.NewMeter. Só o tracing precisa de flag.
func InitSentry(cfg config.Config) error {
	if cfg.SentryDSN == "" {
		return nil
	}
	err := sentry.Init(sentry.ClientOptions{
		Dsn:              cfg.SentryDSN,
		Environment:      cfg.AppEnv,
		Release:          releaseName,
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
