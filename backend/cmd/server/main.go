package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kanban-dev-app/backend/internal/config"
	"kanban-dev-app/backend/internal/database"
	"kanban-dev-app/backend/internal/handler"

	"github.com/getsentry/sentry-go"
	sentryhttp "github.com/getsentry/sentry-go/http"
)

// frontendDist must stay the literal "frontend/dist": the binary's working
// directory in production is the Dockerfile WORKDIR, not backend/.
const frontendDist = "frontend/dist"

// sentryFlushTimeout bounds how long shutdown waits for buffered events.
const sentryFlushTimeout = 5 * time.Second

func main() {
	cfg := config.Load()

	// Sentry comes up FIRST, before the database. Everything below this line can
	// fail in a way worth reporting -- a database that never answers, a
	// migration with broken SQL -- and those failures exit the process, so
	// without Sentry already listening they would vanish into the container log.
	if err := handler.InitSentry(cfg); err != nil {
		slog.Warn("sentry init failed (continuing without it)", "err", err)
	}
	defer sentry.Flush(sentryFlushTimeout)

	// Tee slog to Sentry's structured logs, from Warn up. Info would be mostly
	// request noise; Warn and Error are what someone reading Sentry wants.
	slog.SetDefault(handler.NewAppLogger(cfg.SentryDSN))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Connect to the database (with retry) and run migrations at startup.
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		fatal("failed to connect to database", err)
	}
	defer db.Close()

	if err := database.Migrate(ctx, db); err != nil {
		fatal("migration failed", err)
	}

	mux := http.NewServeMux()

	// Health check — required by the deploy skill (GET /up -> 200).
	mux.HandleFunc("/up", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// JSON API.
	mux.Handle("/api/", handler.NewAPI(db, cfg))

	// Serve the built SPA. Skipped in local dev only, where Vite serves the
	// frontend and proxies /api to this backend. DEV_MODE is never set in a
	// deployed environment — setting it there leaves the app serving 404 on
	// every page.
	if cfg.DevMode {
		slog.Info("dev mode: SPA served by Vite, static handler not registered")
	} else {
		handler.RegisterFrontend(mux, frontendDist)
	}

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: withSentryTracing(mux, cfg),
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	slog.Info("server listening", "port", cfg.Port, "dev_mode", cfg.DevMode)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		fatal("server error", err)
	}
}

// withSentryTracing turns each request into a Sentry transaction and attaches a
// per-request hub, which is what lets handlers report errors onto the right
// trace. Without this middleware TracesSampleRate has nothing to sample: the
// rate decides which transactions are kept, it does not create them.
func withSentryTracing(h http.Handler, cfg config.Config) http.Handler {
	if cfg.SentryDSN == "" {
		return h
	}
	// Repanic keeps Go's default behaviour for a panic that escapes a handler.
	// The API's own recovery runs inside this one and answers 500 before the
	// panic gets here, so nothing is reported twice.
	return sentryhttp.New(sentryhttp.Options{Repanic: true}).Handle(h)
}

// fatal reports, flushes and exits. os.Exit skips deferred calls, so the flush
// has to happen here or the very errors worth reading never leave the process.
func fatal(msg string, err error) {
	slog.Error(msg, "err", err)
	sentry.CaptureException(err)
	sentry.Flush(sentryFlushTimeout)
	os.Exit(1)
}
