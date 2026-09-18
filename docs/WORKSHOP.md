# WORKSHOP — Roteiro e Estado (âncora de contexto)

> Este arquivo é a âncora de contexto da sessão. Se a conversa for
> comprimida, **reler este arquivo primeiro** antes de continuar. Ele
> registra decisões do workshop que NÃO são óbvias no código.

## Objetivo do workshop
ChatOps com o Hermes Agent. Duração: **1h30**.
- 20 min: subir o Hermes + setup do participante.
- ~70 min: fork, publicar, integrar Sentry, corrigir bugs plantados.

## Fluxo ao vivo (o que os participantes fazem)
1. Pedem pro Hermes **forkar** o projeto `kanban-dev-app` (este repo).
2. Instalam o plugin **Cofounder** no projeto (skills da stack + deploy
   Locaweb Cloud / Cloudstack).
3. **Publicam** — pipeline do deploy provisiona as VMs na Locaweb Cloud e
   devolve uma URL `[IP].nip.io`.
4. Criam conta **gratuita no Sentry**; pedem pro Hermes criar a
   **integração com Sentry** (feature nova).
5. **Bug backend**: o app devolve **500 na arrancada**. Pedem pro Hermes
   corrigir → ele cria **PR** e publica a correção.
6. Navegam no board, **criam/movem task** → erros de **frontend** ocorrem.
7. Acessam o **Sentry** e veem as Issues coletadas automaticamente.
8. Pedem **pelo Telegram** pro Hermes verificar as Issues → ele manda uma
   descrição **amigável** do erro no Telegram.
9. Pedem pra corrigir → Hermes corrige e **envia o link do PR** no Telegram.
10. Pedem pra **publicar** a correção dos bugs de frontend.

## RESTRIÇÕES DOS BUGS (crítico — não esquecer)
- **Bug backend NÃO é de lógica.** Tem que ser um **erro concreto de
  sintaxe** — ex.: erro de sintaxe numa **migration** (SQL inválido) que
  faz o Postgres falhar na arrancada → app não sobe → **500**.
- **Bug frontend** é de **runtime JS** ao **criar/mover task** (ex.: campo
  errado na chamada de API, `undefined` não tratado, hook com dependência
  errada) → exceção → **Sentry captura automaticamente**.
- Os bugs são plantados **DEPOIS** que a aplicação está **100% funcional e
  testada**. Primeiro o app funciona; depois plantamos os bugs.

## Decisões de stack / projeto (já aceitas)
- **Nome:** `kanban-dev-app`.
- **Stack:** Go + React + Postgres (padrão Cofounder), sqlc, Kamal.
- **Frontend:** React Router (ssr:false) + **shadcn** + **lucide-react** +
  **Tailwind**. Design bem dev, **tema claro/escuro persistido em
  localStorage**.
- **Login:** **mockado** (usuário demo; `POST /api/dev/login` só com
  DEV_MODE). Sem OAuth/SMTP.
- **Deploy:** a cada **push** dispara a pipeline. **Só faz deploy se houver
  os secrets da Locaweb Cloud no GitHub.** O cliente gera as credenciais no
  painel Locaweb Cloud e cria os secrets no GitHub.
- **Sentry:** backend `sentry-go` (DSN via `SENTRY_DSN`); frontend
  `@sentry/react` (DSN via `NEXT_PUBLIC_SENTRY_DSN` / `VITE_...`).
- **Cada participante configura o próprio Hermes/Telegram** no setup.
- **Repo:** `github.com/fagnerlopes/kanban-dev-app` (private).

## Estado atual (atualizar a cada checkpoint)
- [x] Cofounder instalado no projeto
- [x] Repo criado e pushado (private)
- [x] Toolchain: mise, Go 1.27, sqlc, Node 24, Postgres (container)
- [x] Backend: schema (migration 001), API JSON, /up, gancho Sentry
- [x] Backend: testes Layer 1 (5 testes, PASS)
- [x] Docs: PRD, TASKS, INFRASTRUCTURE, 2 ADRs, WORKSHOP
- [x] Frontend: scaffold React Router + shadcn + Tailwind
- [x] Frontend: tema claro/escuro persistido (toggle ok, sem FOUC)
- [x] Frontend: login mockado
- [x] Frontend: board com drag-and-drop
- [x] Frontend: testes Layer 2 (Vitest, 10 tests PASS)
- [x] Visual check: light + dark (Playwright)
- [x] Commit + push GitHub (master 5417897)
- [ ] Dockerfile multi-stage (frontend + backend)
- [ ] Pipeline de deploy (GHA + Kamal, só com secrets)
- [ ] Bugs plantados (após app funcional)

## Próximo passo
1. Dockerfiles (backend golang:1.27-alpine; frontend node:24-alpine → nginx:alpine,
   build client, serve SPA + proxy /api).
2. Pipeline GHA: push → build/test → **se** secrets Locaweb Cloud existirem →
   deploy via Kamal. Sem secrets, só CI verde (não faz deploy).
3. Plantar os bugs (após tudo verde).
