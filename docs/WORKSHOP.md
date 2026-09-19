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
- [x] Dockerfile multi-stage (frontend + backend)
- [x] Pipeline de deploy (GHA + Kamal, só com secrets)
- [ ] Bugs plantados (após app funcional)

## Estado ao dormir (2026-09-19 ~05:30)

- **Deploy `35423643294`** rodando em background (watch ativo). Provision está
  **criando de verdade** (sem "skipped") após limpar o cache de estado `infra-*`.
- **Causa raiz dos 3 deploys anteriores** (resolvido):
  1. `POSTGRES_PASSWORD` com `-`/`_` quebrava o parse do pgx → senha alfanumérica
     + `normalizeDatabaseURL` (commit `981c63a`).
  2. IP do provision era de **outro projeto** (pool dinâmico do CloudStack).
  3. **Cache de estado** do provision no Actions ficava "stale" após teardown →
     provision "skipped" a criação e o deploy SSH em IPs que não existiam.
     **Fix:** apagar caches `infra-*` do repo (`gh api DELETE .../actions/caches/<id>`).
- **Checklist pra amanhã (se o deploy ainda não subiu):**
  - [ ] `gh run view 35423643294` → se failure, `--log-failed`.
  - [ ] Se "skipped" de novo: limpar caches `infra-*` e re-push.
  - [ ] Se healthy: pegar `web_ip` de `provision-output.json` e abrir
    `https://<web_ip>.nip.io`.
  - [ ] Se app no ar: **plantar bugs** (backend migration syntax + frontend
    runtime) → PR + deploy.
  - [ ] Configurar **Sentry DSN** (hook já no código, DSN faltando).
- **VMs:** as de `191.252.226.176`/`.198` são de OUTRO projeto (não mexer).
- **PAT do gh:** token de admin (muitos escopos) — revisar para escopo mínimo
  (`repo` + `write:packages` + `workflow`) quando possível.

## Próximo passo
1. Confirmar o preview no ar (`https://<web_ip>.nip.io/up` → 200).
2. Plantar os bugs (backend = sintaxe na migration; frontend = runtime no dnd).
