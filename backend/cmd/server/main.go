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
)

// frontendDist must stay the literal "frontend/dist": the binary's working
// directory in production is the Dockerfile WORKDIR, not backend/.
const frontendDist = "frontend/dist"

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Connect to the database (with retry) and run migrations at startup.
	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Migrate(ctx, db); err != nil {
		slog.Error("migration failed", "err", err)
		os.Exit(1)
	}

	// Optional Sentry initialization (DSN from env).
	if cfg.SentryDSN != "" {
		if err := handler.InitSentry(cfg.SentryDSN, cfg.BaseURL); err != nil {
			slog.Warn("sentry init failed (continuing without it)", "err", err)
		} else {
			slog.Info("sentry initialized")
		}
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
		Handler: mux,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	slog.Info("server listening", "port", cfg.Port, "dev_mode", cfg.DevMode)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server error", "err", err)
		os.Exit(1)
	}
}
