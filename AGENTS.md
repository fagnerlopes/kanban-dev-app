<!-- cofounder:begin -->
You are operating as a cofounder. Your FIRST action this session — before responding to or acting on the user's message, even if it is a specific task, question, or feature request — MUST be to use the Skill tool to invoke the `cofounder-playbook` skill, then follow it as your operating instructions for the entire session. Do not begin the user's request until the playbook is loaded. (If the cofounder skills are not installed in this project, first install them by running `mise x node@22 -- npx -y skills add locaweb/cofounder --agent universal claude-code hermes-agent --skill '*' -y`, then invoke the playbook.)
<!-- cofounder:end -->

# Kanban Dev Flow

Kanban board app: Go (stdlib + pgx, sqlc) backend + React SPA frontend (React Router v8, `ssr: false` — pure client, no Node server at runtime). Single root Dockerfile serves the SPA from the Go server on port 80 (health check `GET /up`). Deploy: Locaweb Cloud via Kamal (preview env), Postgres 17 as accessory.

## Dev environment

- Toolchain pinned in `mise.toml` (go 1.27, node 24, sqlc, python 3.14). `go`/`node` are NOT on the default PATH — prefix commands with `mise x --` (e.g. `mise x -- go test ./...`).
- Local Postgres: `supabase/postgres:17.6.1.171` on port 5432 (see `docs/INFRASTRUCTURE.md`).
- Env: copy `.env.example` to `.env` (git-ignored). Key vars: `PORT` (8080 local / 80 in container), `DATABASE_URL`, `BASE_URL` (default `http://localhost:5173`), `DEV_MODE=1` (enables `POST /api/dev/login` — never in production), `SENTRY_DSN`, `APP_ENV` (Sentry's environment label).
- Run backend: `cd backend && mise x -- go run ./cmd/server` (runs migrations at startup; the default columns are seeded by migrations, not by `main.go`).
- Run frontend: `cd frontend && npm run dev` (Vite on 5173; proxies `/api` and `/auth` to `http://localhost:8080`).

## Build & test

- Frontend (in `frontend/`): `npm run dev`, `npm run build`, `npm test` (vitest), `npm run typecheck` (react-router typegen + tsc), `npm run test:watch`.
- Backend (in `backend/`): `mise x -- go test ./...` (handler tests are DB integration tests — they SKIP if `DATABASE_URL` is unset, so a green run may mean nothing; `go run ./cmd/server` also validates the DSN).
- Frontend build output goes to `frontend/build/client` (SPA mode) — the Dockerfile copies it to `frontend/dist`, a sibling of the binary under the image WORKDIR. `frontendDist` in `backend/cmd/server/main.go` must stay the literal `"frontend/dist"`.
- Deploy is automatic: push to `master` triggers `.github/workflows/deploy-preview.yml` (provision + Kamal deploy of the root Dockerfile). It also accepts `workflow_dispatch`, which is what `make deploy` uses — a fresh fork has no commit to push.
- Workshop onboarding lives in `README.md` and `Makefile` (`make setup|deploy|url|sentry|domain|domain-reset|status`); the bootstraps are `scripts/setup-secrets.sh` and `scripts/set-domain.sh`.
- The public hostname comes from the `APP_DOMAIN` repository *variable* (comma-separated for apex+www), resolved in ERB in `config/deploy.preview.yml`. Never hardcode a domain there — every fork would route a name its owner does not control.
- `<web_ip>.nip.io` must stay in `proxy.hosts` **unconditionally**; `APP_DOMAIN` only appends. kamal-proxy routes strictly by `Host` and the deploy health check targets the container directly, so swapping nip.io for an unresolved domain leaves the app unreachable on every address **with a green deploy**. Guarded by `TestDeployConfigsKeepDomainConfigurable` and `TestDeployConfigsAlwaysRouteNipIo`. See `docs/adr/005-dominio-por-variavel-de-ambiente.md`.

## Conventions

- Go module rooted at `backend/` (`kanban-dev-app/backend`); `internal/{config,database,handler}`. SQL in `backend/internal/database/queries/*.sql`, migrations in `backend/internal/database/migrations/`, generated code in `backend/internal/database/sqlc/` — after changing queries/migrations regenerate with `mise x -- sqlc generate` (config: `backend/sqlc.yaml`).
- Frontend is a React Router SPA: routes in `frontend/app/routes/` (registered in `routes.ts`), components in `app/components`, Tailwind v4 + shadcn.
- Commit messages: conventional style (`feat:`, `fix:`, `docs:`, `chore:`, `style:`, `debug:`).

## Pitfalls

- `backend/internal/database/sqlc/` is generated — never hand-edit.
- Migrations run at app startup (embedded runner); `DATABASE_URL` is required and the server exits hard without it.
- `DEV_MODE=1` must never be set in a deployed environment. It does two things: registers `POST /api/dev/login`, and tells the Go server that Vite — not Go — serves the SPA. Setting it in `config/deploy*.yml` makes every HTML page 404 while `/up` and the API stay green, so deploys go through looking healthy. Guarded by `TestDeployConfigsNeverSetDevMode`.
- The demo login is client-side (`frontend/app/hooks/use-auth.ts`) and never calls the API — don't reintroduce a dependency on `/api/dev/login`. See `docs/adr/003-spa-em-producao-e-dev-mode.md`.
- Port 80 belongs to **kamal-proxy** on the deploy VMs. Never install nginx or another reverse proxy; locally the entry point is Vite (5173), not port 80.
- The local Postgres container must be `supabase/postgres:17.6.1.171` — the same image as production, and the one that ships the Cofounder extensions. `postgres:alpine` is not a substitute.
- `e2e/*.png` are local screenshot evidence (git-ignored); `e2e/` only contains ad-hoc Playwright screenshot scripts, not a test suite.
- Deploy secrets: `DATABASE_URL` is composed in the workflow from `POSTGRES_PASSWORD` — Kamal does not expand `$VAR` inside composed values (see commit history for the root cause).
- Sentry Go SDK (v0.49.0) has **no `EnableLogs` option** — the onboarding snippet is stale. Logs and metrics switch on by *using* `sentry.NewLogger` / `sentry.NewMeter`; only tracing needs `EnableTracing` + `TracesSampleRate`, and tracing also needs the `sentryhttp` middleware or the rate has no transactions to sample.
- Never build the app logger by wrapping `slog.Default().Handler()` and then calling `slog.SetDefault`: the default handler writes through the `log` package and `SetDefault` routes `log` back into it, so the two recurse until the `log` mutex self-deadlocks. The process hangs on its first log line with no panic. Use `handler.NewAppLogger`, which constructs its own base handler. Guarded by `TestAppLoggerDoesNotDeadlockAfterSetDefault`.
- `writeErr` reports every non-404 to Sentry; without it a handled 500 left no trace anywhere but the container log. `main.fatal` reports and flushes before `os.Exit`, since deferred flushes never run there.
- The frontend Sentry DSN must come from `GET /api/config`, never from a `VITE_SENTRY_DSN` build variable: Vite inlines `VITE_*` during the Docker build while Kamal injects secrets only at runtime, so the bundle would ship an empty DSN with a green pipeline. See `docs/adr/004-sentry-dsn-em-runtime.md`.
- Never run the Sentry wizard (`npx @sentry/wizard`) here: it wires `@sentry/vite-plugin`, which needs a `SENTRY_AUTH_TOKEN` at Docker build time (same trap), hardcodes the DSN into source, and is interactive. Source maps are instead emitted by the build (`build.sourcemap: true`) and served publicly by `RegisterFrontend`; Sentry fetches them from the `sourceMappingURL`. Guarded by `TestViteBuildEmitsSourceMaps` and `TestFrontendServesSourceMaps`.
