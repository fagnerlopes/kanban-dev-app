package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"kanban-dev-app/backend/internal/config"
	"kanban-dev-app/backend/internal/database"
	"kanban-dev-app/backend/internal/handler"
)

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

	// Serve the built frontend (SPA) in non-dev mode.
	registerFrontend(mux, cfg)

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

// registerFrontend serves the built SPA from the frontendDist directory.
// The SPA shell is detected (index.html, or __spa-fallback.html when the build
// prerenders "/"). Real files are served by http.FileServer (correct
// Content-Type); unmatched routes fall back to the SPA shell. This handler is
// only registered in non-dev mode — in dev, Vite serves the frontend and
// proxies /api to this backend.
//
// frontendDist must be the literal "frontend/dist": the binary's working
// directory in production is the Dockerfile WORKDIR, not backend/.
func registerFrontend(mux *http.ServeMux, cfg config.Config) {
	if cfg.DevMode {
		return
	}
	frontendDist := "frontend/dist"
	if _, err := os.Stat(frontendDist); err != nil {
		slog.Warn("frontend dist not found — SPA routes will return 404", "path", frontendDist)
		return
	}

	spaShell := "index.html"
	if _, err := os.Stat(filepath.Join(frontendDist, "__spa-fallback.html")); err == nil {
		spaShell = "__spa-fallback.html"
	}

	fs := http.FileServer(http.Dir(frontendDist))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") || strings.HasPrefix(r.URL.Path, "/auth/") {
			http.NotFound(w, r)
			return
		}
		if r.URL.Path == "/" {
			fs.ServeHTTP(w, r)
			return
		}
		target := filepath.Join(frontendDist, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(target); err == nil {
			if !info.IsDir() {
				fs.ServeHTTP(w, r) // real asset; FileServer sets Content-Type
				return
			}
			page := filepath.Join(target, "index.html")
			if _, err := os.Stat(page); err == nil {
				http.ServeFile(w, r, page)
				return
			}
		}
		http.ServeFile(w, r, filepath.Join(frontendDist, spaShell))
	})
}
