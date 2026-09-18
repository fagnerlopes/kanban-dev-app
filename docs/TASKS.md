# Tasks — Kanban Dev Flow

Tracker de desenvolvimento. Atualizado a cada sessão.

| Task | Status | Reason/Notes |
|------|--------|--------------|
| Estrutura do projeto (Cofounder + toolchain) | Done | mise, Go 1.27, sqlc, Node 24, Postgres |
| Repo no GitHub (private) | Done | fagnerlopes/kanban-dev-app |
| Backend: schema do kanban (migration 001) | Done | columns + tasks + seed das 5 colunas |
| Backend: API JSON (board, CRUD de tasks) | Done | testado via curl |
| Backend: health check /up | Done | 200 ok |
| Backend: gancho Sentry (InitSentry + recover) | Done | dispara 500 + captura panic |
| Docs: PRD, TASKS, INFRASTRUCTURE, ADR | Done | PRD + TASKS + INFRA + 2 ADRs |
| Backend: testes Layer 1 (handlers Go) | Done | 5 testes, table-driven, HTTP real, todos PASS |
| Frontend: scaffold React Router + shadcn + Tailwind | Pending | ssr:false, proxy dev |
| Frontend: tema claro/escuro persistido | Pending | localStorage |
| Frontend: login mockado | Pending | POST /api/dev/login |
| Frontend: board com drag-and-drop | Pending | criar/mover/remover task |
| Frontend: testes Layer 2 (Vitest) | Pending | componentes interativos |
| Dockerfile multi-stage | Pending | front + back + runtime Alpine |
| Pipeline de deploy (GHA + Kamal) | Pending | só com secrets Locaweb |
| Bugs plantados (backend migration + frontend) | Pending | após app funcional |
