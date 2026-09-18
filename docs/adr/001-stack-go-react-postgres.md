# 001 - Stack Go + React + Postgres (padrão Cofounder)

**Status:** Accepted

## Context

O projeto deve seguir a stack prescrita pelo Cofounder (Locaweb) para
alinhar com a pipeline de deploy (Kamal + Cloudstack) e com as skills do
plugin. O backend precisa ser um binário único que sirva a API e o
frontend estático.

## Decision

- **Backend:** Go stdlib `net/http` + `pgx/v5` (via `stdlib` para
  `database/sql`) + `sqlc` para geração de queries.
- **Frontend:** React + TypeScript em React Router framework mode
  (`ssr: false`), shadcn/ui, Tailwind CSS, Vite.
- **Banco:** PostgreSQL (supabase/postgres), migrations forward-only
  embutidas via `go:embed`, aplicadas na arrancada com retry.
- **Deploy:** imagem Docker única (front + back), orquestrada por Kamal na
  Locaweb Cloud.

## Rationale

- Alinhamento com o plugin Cofounder e com o deploy Locaweb (mesma stack do
  LeanTrack).
- `sqlc` garante type-safety das queries sem ORM.
- Binário único simplifica o deploy e o health check.

## Trade-offs

**Pros:**
- Type safety no banco (sqlc).
- Deploy simples (um container).
- Ecossistema maduro e documentado no plugin.

**Cons:**
- Dois languages (Go + TS) — mais toolchain.
- Sem SSR: páginas públicas dependerão de prerender (não crítico para um app
  atrás de login).

## Alternatives Considered

- **Node full-stack (Next.js):** descartado — o Cofounder prescreve binário
  Go único e não usa Node server em runtime.
- **SQLite:** descartado — o deploy Locaweb provisiona Postgres como VM;
  alinhar local e produção com a mesma imagem.
