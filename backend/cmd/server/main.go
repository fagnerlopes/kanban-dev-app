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

// ARMADILHA: tem de ser o literal "frontend/dist". Em produção o diretório de
// trabalho é o WORKDIR da imagem, não backend/ — qualquer outro caminho passa
// nos testes locais e devolve 404 em todas as páginas depois do deploy.
const frontendDist = "frontend/dist"

const sentryFlushTimeout = 5 * time.Second

func main() {
	cfg := config.Load()

	// Antes do banco: falhas de arranque saem com os.Exit e precisam de um
	// Sentry já escutando para não sumirem.
	if err := handler.InitSentry(cfg); err != nil {
		slog.Warn("sentry init failed (continuing without it)", "err", err)
	}
	defer sentry.Flush(sentryFlushTimeout)

	slog.SetDefault(handler.NewAppLogger(cfg.SentryDSN))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		fatal("failed to connect to database", err)
	}
	defer db.Close()

	if err := database.Migrate(ctx, db); err != nil {
		fatal("migration failed", err)
	}

	mux := http.NewServeMux()

	// ARMADILHA: /up e a sonda de vida do kamal-proxy e nao pode consultar o
	// banco. Um blip do Postgres tiraria o app da rota e reprovaria deploys.
	// O check profundo (banco incluso) e GET /api/health.
	mux.HandleFunc("/up", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.Handle("/api/", handler.NewAPI(db, cfg))

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

// withSentryTracing cria a transação de cada requisição e anexa o hub. Sem o
// middleware, TracesSampleRate não tem o que amostrar.
func withSentryTracing(h http.Handler, cfg config.Config) http.Handler {
	if cfg.SentryDSN == "" {
		return h
	}
	return sentryhttp.New(sentryhttp.Options{Repanic: true}).Handle(h)
}

// fatal reporta e dá flush antes de sair: os.Exit não roda defers.
func fatal(msg string, err error) {
	slog.Error(msg, "err", err)
	sentry.CaptureException(err)
	sentry.Flush(sentryFlushTimeout)
	os.Exit(1)
}
