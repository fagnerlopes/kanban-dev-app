# Tasks — Kanban Dev Flow

Tracker de desenvolvimento. Atualizado a cada sessão.

| Task | Status | Reason/Notes |
|------|--------|--------------|
| Estrutura do projeto (Cofounder + toolchain) | Done | mise, Go 1.27, sqlc, Node 24, Postgres |
| Repo no GitHub | Done | fagnerlopes/kanban-dev-app (público) |
| Backend: schema do kanban (migration 001) | Done | columns + tasks + seed das 5 colunas |
| Backend: API JSON (board, CRUD de tasks) | Done | testado via curl e testes de integração |
| Backend: health check /up | Done | 200 ok |
| Backend: gancho Sentry (InitSentry + recover) | Done | dispara 500 + captura panic |
| Backend: testes Layer 1 (handlers Go) | Done | handlers + config + SPA, todos PASS |
| Frontend: scaffold React Router + shadcn + Tailwind | Done | ssr:false, proxy dev |
| Frontend: tema claro/escuro persistido | Done | localStorage, sem FOUC |
| Frontend: login mockado | Done | sessão demo client-side (ADR-003) |
| Frontend: board com drag-and-drop | Done | criar/mover/remover task |
| Frontend: testes Layer 2 (Vitest) | Done | 15 testes (componentes + useAuth) |
| Dockerfile multi-stage | Done | node:24-alpine + golang:1-alpine + distroless |
| Pipeline de deploy (GHA + Kamal) | Done | provision + Kamal, dispara no push da master |
| **Fix: SPA retornando 404 em produção** | Done | `DEV_MODE=1` no deploy desligava o servidor de estáticos (ADR-003) |
| **Fix: login dependia de rota só-de-dev** | Done | sessão demo agora é client-side |
| **Fix: banco local com imagem errada** | Done | era `postgres:17-alpine`, agora `supabase/postgres:17.6.1.171` |
| **Fix: nginx indevido na máquina de dev** | Done | removido; quem faz proxy da porta 80 é o kamal-proxy |
| **Fix: board transbordando em 1280px** | Done | colunas dividem a largura; rolagem só quando não cabe |
| Limpeza: scripts e arquivos fora do padrão | Done | `pgxtest_main.go`, `.dockerignore`, favicon de template |
| README com o passo a passo do participante | Done | fork → Actions → Locaweb Cloud → `make setup` → `make deploy` → Sentry |
| `make setup` / `deploy` / `url` / `sentry` / `status` | Done | `Makefile` + `scripts/setup-secrets.sh`, idempotentes |
| `workflow_dispatch` no deploy | Done | fork novo não tem push para disparar a pipeline |
| Backend: `GET /api/config` (DSN em runtime) | Done | substitui `VITE_SENTRY_DSN` — ver [ADR-004](adr/004-sentry-dsn-em-runtime.md) |
| Backend: `APP_ENV` como ambiente do Sentry | Done | antes ia `BASE_URL`, o que virava URL no facet do Sentry |
| Source maps servidos pelo app (Sentry lê pela URL) | Done | `build.sourcemap: true` no Vite; sem `SENTRY_AUTH_TOKEN`, sem upload. Guardado por `TestViteBuildEmitsSourceMaps` e `TestFrontendServesSourceMaps` |
| README: aviso de bloqueador de anúncios | Done | ad-blocker derruba só os eventos de navegador — sintoma confuso |
| Crédito "Feito com Cofounder e Locaweb Cloud" | Done | componente `made-with.tsx`, no rodapé do login e do board; links com contraste ≥7:1 |
| Domínio personalizado por variável (`APP_DOMAIN`) | Done | `make domain` confere o DNS antes de publicar; fallback `nip.io`. Ver [ADR-005](adr/005-dominio-por-variavel-de-ambiente.md) |
| Sentry: integração no frontend | Pending | **de propósito** — é a feature que o Hermes cria ao vivo (`@sentry/react` lendo `/api/config`, **sem o wizard**) |
| Sentry: configurar DSN real | Done | secret gravado via `make sentry`; evento de teste confirmado |
| Sentry: reportar 500 tratados (`writeErr`) | Done | antes só `panic` chegava; um 500 de query quebrada era invisível |
| Sentry: reportar falhas de arranque | Done | banco fora do ar / migration quebrada agora reportam e dão flush antes do `os.Exit` |
| Sentry: logs (`slog` → structured logs, Warn+) | Done | ponte em `sentryslog.go`; sem `EnableLogs`, que não existe no SDK Go |
| Sentry: tracing (`sentryhttp` + sample rate) | Done | `SENTRY_TRACES_SAMPLE_RATE`, default 1 |
| Sentry: Application Metrics | Done | `kanban.task.{created,updated,deleted}` |
| **Fix: deadlock ao instalar o logger** | Done | embrulhar `slog.Default().Handler()` + `SetDefault` trava o processo na 1ª linha de log, sem panic |
| `GET /api/health` para monitoramento externo | Done | 200/503 com check de banco; separado do `/up` do kamal-proxy |
| Ícone do GitHub no header (login + board) | Done | link vem de `repo_url` em `/api/config`, então cada fork aponta para si mesmo |
| Seed de cards de exemplo (migration 003) | Done | 10 Backlog / 3 To Do / 4 In Dev / 2 Review / 4 Done; idempotente e não toca em cards existentes |
| Bugs plantados (backend migration + frontend) | Pending | só depois do app 100% funcional — ver `WORKSHOP.md` |

## Dívida conhecida (não bloqueia o workshop)

| Item | Nota |
|------|------|
| `PATCH /api/tasks/{id}` ignora `description` | A query `UpdateTask` não atualiza a descrição; hoje nenhuma tela edita esse campo, então não aparece. |
| Bug de backend planejado derruba o deploy, não gera 500 | O `WORKSHOP.md` prevê erro de sintaxe numa migration. Como as migrations rodam na subida, o processo sai com `os.Exit(1)`, o Kamal não promove a versão e o **deploy falha** — o participante vê pipeline vermelha, não um 500. Decidir na hora de plantar: ou aceitar isso como o roteiro, ou trocar por um erro que só dispara ao servir uma rota. |
| Senha do Postgres local | O `DATABASE_URL` do `.env` usa uma senha de 3 caracteres. Só afeta a máquina local (o deploy usa o secret `POSTGRES_PASSWORD`), mas vale trocar. |
