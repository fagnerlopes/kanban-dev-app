# Infrastructure — Kanban Dev Flow

Serviços que o app depende, além do binário Go.

| Name | Image | Local Port | Env Var | Type |
|------|-------|-----------|---------|------|
| db | supabase/postgres:17.6.1.171 | 5432 | DATABASE_URL | backend |

## Notas de deploy

- **db** — Postgres é o storage primário. Em deploy, provisionado como VM
  dedicada na Locaweb Cloud (Kamal); hostname `db:5432` na rede interna
  Cloudstack.
- **Sentry** — serviço externo SaaS (Sentry). Não é um accessory de VM; a
  integração é feita via `SENTRY_DSN` (backend) e `NEXT_PUBLIC_SENTRY_DSN`
  (frontend). DSN vazio = Sentry desativado.
