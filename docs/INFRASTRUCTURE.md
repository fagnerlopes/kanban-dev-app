# Infrastructure — Kanban Dev Flow

Serviços que o app depende, além do binário Go.

| Name | Image | Local Port | Env Var | Type |
|------|-------|-----------|---------|------|
| db | supabase/postgres:17.6.1.171 | 5432 | DATABASE_URL | backend |

## Deploy (Locaweb Cloud via Kamal)

- **1 web VM** (plan `small`, 20 GB disk) — roda a imagem única do repo
  (Go server + SPA embutida), escuta na porta 80, health check em `GET /up`.
  URL: `https://<web_ip>.nip.io` (TLS Let's Encrypt via kamal-proxy).
- **1 accessory `db`** (plan `small`, 20 GB disk) — Postgres
  `supabase/postgres:17.6.1.171`. Em rede interna CloudStack, hostname
  determinístico `db:5432` (DNS interno). Data em `/data/pgdata`.
- **Sem workers** — o app não tem jobs em background.

### Secrets (GitHub Actions)

| Secret | Origem | Uso |
|--------|--------|-----|
| `CLOUDSTACK_API_KEY` / `CLOUDSTACK_SECRET_KEY` | painel Locaweb Cloud | provision de VMs (o usuário cria no painel e grava no GitHub) |
| `SSH_PRIVATE_KEY` | gerada no setup | deploy Kamal + debug SSH |
| `POSTGRES_PASSWORD` | gerada no setup | senha do Postgres (derivada em `DATABASE_URL`) |
| `SENTRY_DSN` | painel Sentry (workshop) | reporting de erros (vazio = desativado) |

`DATABASE_URL` é **derivada** de `POSTGRES_PASSWORD` no `.kamal/secrets.preview`
(não é um GitHub Secret).

### Gatilho de deploy

- **`push` na `master`** → `deploy-preview.yml` (provision + Kamal).
- O deploy **só acontece se os secrets `CLOUDSTACK_API_KEY`/`CLOUDSTACK_SECRET_KEY`
  existirem** no GitHub. Sem eles, o workflow falha no provision (sem provisionar
  nada). Sem `SENTRY_DSN`, o app sobe com Sentry desativado.
- **`workflow_dispatch`** → `teardown-preview.yml` (destrói o ambiente).

## Notas

- **Sentry** — serviço externo SaaS. Não é um accessory de VM; a integração é
  feita via `SENTRY_DSN` (backend, `sentry-go`) e `NEXT_PUBLIC_SENTRY_DSN`
  (frontend, `@sentry/react`). DSN vazio = Sentry desativado.
- Migrations rodam **no startup do container** (web VM única, sem race).
