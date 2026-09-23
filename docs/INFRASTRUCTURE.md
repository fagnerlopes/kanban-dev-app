# Infrastructure — Kanban Dev Flow

Serviços que o app depende, além do binário Go.

| Name | Image | Local Port | Env Var | Type |
|------|-------|-----------|---------|------|
| db | supabase/postgres:17.6.1.171 | 5432 | DATABASE_URL | backend |

O container local **precisa** usar essa mesma imagem (`<repo>-db`), e não
`postgres:alpine`: é ela que traz as extensões do padrão Cofounder (pgmq,
pg_cron, pgroonga, pgvector, pg_jsonschema, PostGIS) e é a que roda em
produção.

## Deploy (Locaweb Cloud via Kamal)

- **1 web VM** (plan `small`, 20 GB disk) — roda a imagem única do repo
  (Go server + SPA embutida), escuta na porta 80, health check em `GET /up`.
  URL: `https://<web_ip>.nip.io` — **sempre roteado** — mais o domínio de
  `APP_DOMAIN` quando definida (TLS Let's Encrypt via kamal-proxy). O
  certificado é emitido por desafio **HTTP-01**, então o domínio só atende
  depois que o DNS aponta para a VM; `make domain` confere isso antes.
  Atenção: o kamal-proxy roteia por cabeçalho `Host` e o health check do deploy
  fala direto com o container, então um domínio mal configurado **não** reprova
  o deploy — ele passa verde e o nome apenas não responde. É por isso que o
  `nip.io` nunca sai da lista (`TestDeployConfigsAlwaysRouteNipIo`).
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
| `SENTRY_DSN` | painel Sentry (workshop) | reporting de erros do backend **e** do navegador (vazio = desativado) |

Os quatro primeiros são criados de uma vez por `make setup`
(`scripts/setup-secrets.sh`); o `SENTRY_DSN` por `make sentry`. Ver o
[README](../README.md) para o passo a passo do participante.

`DATABASE_URL` é **derivada** de `POSTGRES_PASSWORD` no `.kamal/secrets.preview`
(não é um GitHub Secret).

### Variables (GitHub Actions)

| Variable | Origem | Uso |
|----------|--------|-----|
| `APP_DOMAIN` | `make domain` | domínio(s) **adicionais** do app — o `nip.io` é roteado de qualquer forma. Aceita lista separada por vírgula (apex + www); o primeiro item é canônico e vira `BASE_URL` |

É **variable**, não secret: domínio é informação pública, e como secret ficaria
ilegível no painel. Resolvida em ERB no `config/deploy.preview.yml`. Ver
[ADR-005](adr/005-dominio-por-variavel-de-ambiente.md).

### Gatilho de deploy

- **`push` na `master`** → `deploy-preview.yml` (provision + Kamal).
- **`workflow_dispatch`** → `deploy-preview.yml` também. Um fork recém-criado
  não tem commit novo para empurrar, então sem esse gatilho o participante não
  teria como publicar (é o que o `make deploy` usa).
- O deploy **só acontece se os secrets `CLOUDSTACK_API_KEY`/`CLOUDSTACK_SECRET_KEY`
  existirem** no GitHub. Sem eles, o workflow falha no provision (sem provisionar
  nada). Sem `SENTRY_DSN`, o app sobe com Sentry desativado.
- Em forks, o GitHub desativa Actions por padrão: é preciso clicar uma vez em
  *"I understand my workflows, go ahead and enable them"* na aba Actions.
- **`workflow_dispatch`** → `teardown-preview.yml` (destrói o ambiente).

## Notas

- **Sentry** — serviço externo SaaS. Não é um accessory de VM. No backend a
  integração já existe via `SENTRY_DSN` (`sentry-go`); DSN vazio = desativado.
  No frontend ela ainda **não** foi implementada — é de propósito um passo do
  workshop, feito pelo Hermes ao vivo.
- **O que o backend manda para o Sentry hoje:** erros 500 (capturados em
  `writeErr`), panics (`recoverWithSentry`), falhas de arranque — banco fora do
  ar, migration quebrada — que agora são reportadas **antes** do `os.Exit`;
  logs `slog` de `Warn` para cima (ponte em `sentryslog.go`); traces de todas as
  requisições (`sentryhttp` + `SENTRY_TRACES_SAMPLE_RATE`); e contadores
  `kanban.task.{created,updated,deleted}`.
- **`EnableLogs` não existe** no SDK Go (v0.49.0, a mais recente). O snippet de
  onboarding do Sentry mostra essa opção, mas ela virou `DisableLogs` e depois
  foi removida: logs e métricas são ligados pelo **uso** das APIs
  (`sentry.NewLogger`, `sentry.NewMeter`). Só o tracing ainda precisa de flag.
- **Nunca embrulhar `slog.Default().Handler()` e chamar `slog.SetDefault`.**
  O handler padrão do slog escreve pelo pacote `log`, e o `SetDefault` manda o
  pacote `log` de volta para o handler padrão — os dois se chamam até o mutex do
  `log` travar contra si mesmo. O processo congela na primeira linha de log, sem
  panic e sem mensagem. Guardado por
  `TestAppLoggerDoesNotDeadlockAfterSetDefault`.
- **O DSN do frontend NÃO vem de `VITE_SENTRY_DSN`.** O Vite congela variáveis
  `VITE_*` durante o build da imagem, e o Kamal só entrega secrets em runtime —
  o secret ficaria vazio no bundle com a pipeline verde. O backend serve o DSN
  em **`GET /api/config`** (`sentry_dsn`, `environment`, `release`) e a SPA lê
  dali. Ver [ADR-004](adr/004-sentry-dsn-em-runtime.md).
- **`REPO_URL`** (`env.clear`, derivada de `REPO_FULL` no workflow) alimenta o
  `repo_url` de `/api/config`, que é o destino do ícone do GitHub no header.
  Pelo mesmo motivo do domínio: fixar a URL faria todo fork apontar para o
  repositório de origem.
- **`APP_ENV`** (`env.clear` de cada `config/deploy.<env>.yml`) é o rótulo de
  ambiente que o Sentry usa para agrupar as issues. Coberto por
  `TestDeployConfigsSetAppEnv`.
- **Source maps** saem do build (`build.sourcemap: true` em
  `frontend/vite.config.ts`), são servidos como arquivos comuns por
  `RegisterFrontend` e o Sentry os busca pela URL do `sourceMappingURL`. Não há
  `SENTRY_AUTH_TOKEN` nem etapa de upload — e não deve haver: esse token seria
  necessário no **build da imagem**, repetindo a armadilha do `VITE_SENTRY_DSN`.
  Guardado por `TestViteBuildEmitsSourceMaps` e `TestFrontendServesSourceMaps`.
- **Duas rotas de saúde, com papéis diferentes.** `GET /up` é a sonda de vida
  do kamal-proxy: não consulta o banco, porque um blip do Postgres tiraria o app
  da rota e reprovaria deploys. `GET /api/health` é o check profundo (inclui
  `PingContext` no banco), devolve 503 quando algo está fora e é a rota feita
  para monitoramento externo. Falhas dele **não** vão ao Sentry: uma sondagem
  por minuto com o banco fora geraria um evento por minuto.
- Migrations rodam **no startup do container** (web VM única, sem race). A 003
  popula o quadro com 23 cards de exemplo: só adiciona, nunca altera ou remove,
  e o `NOT EXISTS` por (coluna, título) a torna idempotente mesmo se reaplicada
  à mão.
- **Sem reverse proxy próprio.** Quem termina TLS e roteia a porta 80/443 é o
  **kamal-proxy**, provisionado pelo Kamal na web VM. Instalar nginx (ou
  qualquer outro proxy) na frente disso está fora do padrão e conflita com o
  kamal-proxy pela porta 80.
