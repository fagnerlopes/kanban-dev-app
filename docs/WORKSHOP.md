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
- **Login:** **mockado** (usuário demo, sessão criada no cliente e guardada
  no `localStorage`). `POST /api/dev/login` existe só com `DEV_MODE=1`, para
  testes locais — o frontend não depende dela. Sem OAuth/SMTP.
- **Deploy:** a cada **push** dispara a pipeline. **Só faz deploy se houver
  os secrets da Locaweb Cloud no GitHub.** O cliente gera as credenciais no
  painel Locaweb Cloud e cria os secrets no GitHub.
- **Sentry:** backend `sentry-go` (DSN via `SENTRY_DSN`); frontend
  `@sentry/react` lendo o DSN de **`GET /api/config`**, não de
  `VITE_SENTRY_DSN`. O Vite congela variáveis `VITE_*` no build da imagem e o
  Kamal só entrega secrets em runtime: o secret chegaria vazio no bundle com a
  pipeline verde. Um único secret atende os dois lados. Ver
  [ADR-004](adr/004-sentry-dsn-em-runtime.md).
- **Nada de wizard do Sentry.** `npx @sentry/wizard` exige
  `SENTRY_AUTH_TOKEN` no build da imagem, grava o DSN no código (cada
  participante tem o seu) e é interativo. O Hermes integra na mão — o texto do
  pedido está no README, passo 6.3.
- **Source maps** vêm do build (`build.sourcemap: true`) e são servidos pelo
  próprio app; o Sentry os busca pela URL. Sem eles a Issue de frontend chega
  minificada e não dá para pedir correção ao Hermes.
- **Avisar a turma para desligar o bloqueador de anúncios.** Ele barra
  `*.ingest.sentry.io` e derruba **só** os eventos de navegador — os do
  servidor continuam chegando, o que faz o sintoma parecer um bug da
  integração.
- **Cada participante configura o próprio Hermes/Telegram** no setup.
- **Repo:** `github.com/fagnerlopes/kanban-dev-app` (público).

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
- [x] Crédito "Feito com Cofounder e Locaweb Cloud" (login + board)
- [x] Domínio personalizado opcional via `APP_DOMAIN` (`make domain`)
- [ ] Bugs plantados (após app funcional)

## Estado em 2026-09-19 (sessão de correção)

O app **está funcional**. O que estava quebrado e foi corrigido nesta sessão:

1. **A interface retornava 404 em produção.** `config/deploy.preview.yml`
   setava `DEV_MODE: "1"`, e no servidor Go essa flag significa "o Vite serve
   o frontend" — ou seja, o Go parava de servir a SPA. `/up` e `/api/board`
   continuavam saudáveis, então três deploys passaram "verdes" com o app
   inutilizável. Ver [ADR-003](adr/003-spa-em-producao-e-dev-mode.md).
2. **O login dependia de `POST /api/dev/login`**, rota que só existe com
   `DEV_MODE=1` — foi o motivo de terem ligado a flag. Agora a sessão demo é
   criada no cliente e não chama a rede.
3. **nginx indevido na máquina de dev** (`hermes-lab`), instalado às 11:27 de
   19/09 com um site `kanban-dev` fazendo proxy da porta 80 para o Vite.
   Não faz parte do padrão Cofounder — quem termina TLS e ocupa a porta 80 é
   o **kamal-proxy**, nas VMs de deploy. Serviço parado, desabilitado e site
   removido (o pacote segue instalado, inativo).
4. **Banco local com a imagem errada** (`postgres:17-alpine` em vez de
   `supabase/postgres:17.6.1.171`) — sem as extensões do padrão e diferente
   da produção. Container recriado; o antigo ficou preservado como
   `kanban-db-old-alpine`.
5. **Board transbordando** em 1280px (5 colunas de largura fixa cortavam a
   primeira). As colunas agora dividem a largura e só rolam quando não cabem.

**Correção de nota antiga:** as VMs `191.252.226.176` (web) e `191.252.226.198`
(db) **são deste projeto**. A nota anterior dizia que eram de outro projeto —
estava errada: os deploys mais recentes foram para elas e a API responde com o
board deste app.

**Verificado nesta sessão:** imagem do container rodando com `PORT=80`, servindo
`/`, rotas da SPA, assets com `Content-Type` correto, `/api/board`, e
`POST /api/dev/login` corretamente **ausente** (404) fora do modo dev.

## Estado em 2026-09-22 (sessão de onboarding do participante)

Foco: tornar o caminho **fork → publicar → Sentry** fácil o bastante para caber
em 30 minutos sem depuração ao vivo.

1. **Criado o `README.md`** — o repositório não tinha nenhuma porta de entrada.
   Passo a passo completo do participante, do fork ao ChatOps.
2. **`make setup` / `deploy` / `url` / `sentry` / `status`** (`Makefile` +
   `scripts/setup-secrets.sh`). O setup gera a chave SSH e a senha do Postgres,
   pede as duas chaves da Locaweb Cloud com digitação invisível, e é
   idempotente (`--force` recria).
3. **`workflow_dispatch` no `deploy-preview.yml`.** Um fork novo não tem commit
   para empurrar — sem esse gatilho, "publicar" exigiria inventar um commit.
4. **Armadilha do `VITE_SENTRY_DSN` eliminada.** Era o maior risco do roteiro:
   o secret ficaria vazio no bundle e a pipeline passaria verde. O backend
   agora serve o DSN em `GET /api/config`. Ver
   [ADR-004](adr/004-sentry-dsn-em-runtime.md).
5. **`APP_ENV` vira o ambiente do Sentry** (antes ia `BASE_URL`, virando uma
   URL no facet). Coberto por `TestDeployConfigsSetAppEnv`.

A integração do `@sentry/react` no frontend segue **intencionalmente por
fazer** — é a feature que o Hermes cria ao vivo, no passo 4 do fluxo.

## Próximo passo
1. Configurar o `SENTRY_DSN` real no repo original e validar o fluxo ponta a
   ponta uma vez (criar projeto no Sentry, `make sentry`, ver o evento chegar).
2. Plantar os bugs (backend = sintaxe na migration; frontend = runtime no dnd).
   **Antes disso**, reler a dívida anotada em `TASKS.md`: o bug de migration
   derruba o deploy em vez de gerar 500.
