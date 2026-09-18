# 002 - Login mockado para a demo

**Status:** Accepted

## Context

O app é uma demo de workshop. O roteiro não inclui fluxo real de
autenticação (OAuth/magic link exigiriam SMTP + Google Cloud, fora do tempo
do workshop). O Cofounder permite self sign-in para prototipagem, e exige um
endpoint de login de dev (`POST /api/dev/login`, só com `DEV_MODE=1`) para
que testes e visual check alcancem rotas autenticadas.

## Decision

- Autenticação **mockada**: um único usuário demo. O frontend guarda um
  "token" em memória/localStorage após `POST /api/dev/login`.
- O endpoint de dev login só é registrado quando `DEV_MODE=1` (guarda na
  hora do registro da rota). Em produção `DEV_MODE` nunca é setado — a rota
  não existe.
- Sem hash de senha, sem sessão real no banco.

## Rationale

- Mantém o foco do workshop no ChatOps + Sentry, não em auth.
- Satisfa o requisito do Cofounder de ter rota autenticada testável.

## Trade-offs

**Pros:**
- Zero configuração externa.
- Testes/visual check funcionam sem credenciais reais.

**Cons:**
- Inseguro para produção (aceitável: a demo nunca roda em produção real).

## Alternatives Considered

- **Magic link / Google Auth:** descartado — exige SMTP + console Google,
  fora do escopo e do tempo do workshop.
- **Sem login (app público):** descartado — o Cofounder exige rotas
  autenticadas para o visual check; manter o login mockado.
