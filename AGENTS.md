<!-- cofounder:begin -->
You are operating as a cofounder. Your FIRST action this session — before responding to or acting on the user's message, even if it is a specific task, question, or feature request — MUST be to use the Skill tool to invoke the `cofounder-playbook` skill, then follow it as your operating instructions for the entire session. Do not begin the user's request until the playbook is loaded. (If the cofounder skills are not installed in this project, first install them by running `mise x node@22 -- npx -y skills add locaweb/cofounder --agent universal claude-code hermes-agent --skill '*' -y`, then invoke the playbook.)
<!-- cofounder:end -->

# Kanban Dev Flow

Kanban board app: Go (stdlib + pgx, sqlc) backend + React SPA frontend (React Router v8, `ssr: false` — pure client, no Node server at runtime). Single root Dockerfile serves the SPA from the Go server on port 80 (health check `GET /up`). Deploy: Locaweb Cloud via Kamal (preview env), Postgres 17 as accessory.

## Dev environment

- Toolchain pinned in `mise.toml` (go 1.27, node 24, sqlc, python 3.14). `go`/`node` are NOT on the default PATH — prefix commands with `mise x --` (e.g. `mise x -- go test ./...`).
- Local Postgres: `supabase/postgres:17.6.1.171` on port 5432 (see `docs/INFRASTRUCTURE.md`).
- Env: copy `.env.example` to `.env` (git-ignored). Key vars: `PORT` (8080 local / 80 in container), `DATABASE_URL`, `BASE_URL` (default `http://localhost:5173`), `DEV_MODE=1` (enables `POST /api/dev/login` — never in production), `SENTRY_DSN`.
- Run backend: `cd backend && mise x -- go run ./cmd/server` (runs migrations + seeds default columns at startup).
- Run frontend: `cd frontend && npm run dev` (Vite on 5173; proxies `/api` and `/auth` to `http://localhost:8080`).

## Build & test

- Frontend (in `frontend/`): `npm run dev`, `npm run build`, `npm test` (vitest), `npm run typecheck` (react-router typegen + tsc), `npm run test:watch`.
- Backend (in `backend/`): `mise x -- go test ./...` (handler tests are DB integration tests — they SKIP if `DATABASE_URL` is unset, so a green run may mean nothing; `go run ./cmd/server` also validates the DSN).
- Frontend build output goes to `frontend/build/client` (SPA mode) — the Dockerfile copies it from there; don't expect `dist/`.
- Deploy is automatic: push to `master` triggers `.github/workflows/deploy-preview.yml` (provision + Kamal deploy of the root Dockerfile).

## Conventions

- Go module rooted at `backend/` (`kanban-dev-app/backend`); `internal/{config,database,handler}`. SQL in `backend/internal/database/queries/*.sql`, migrations in `backend/internal/database/migrations/`, generated code in `backend/internal/database/sqlc/` — after changing queries/migrations regenerate with `mise x -- sqlc generate` (config: `backend/sqlc.yaml`).
- Frontend is a React Router SPA: routes in `frontend/app/routes/` (registered in `routes.ts`), components in `app/components`, Tailwind v4 + shadcn.
- Commit messages: conventional style (`feat:`, `fix:`, `docs:`, `chore:`, `style:`, `debug:`).

## Pitfalls

- `backend/internal/database/sqlc/` is generated — never hand-edit.
- Migrations run at app startup (embedded runner); `DATABASE_URL` is required and the server exits hard without it.
- `DEV_MODE=1` must never be set in production (exposes the dev login route).
- `e2e/*.png` are local screenshot evidence (git-ignored); `e2e/` only contains ad-hoc Playwright screenshot scripts, not a test suite.
- Deploy secrets: `DATABASE_URL` is composed in the workflow from `POSTGRES_PASSWORD` — Kamal does not expand `$VAR` inside composed values (see commit history for the root cause).
- `backend/pgxtest_main.go` is a standalone DB-ping debug script (`package main` at module root) — it's a scratch tool, not part of the server.
