# 004 - Sentry sem build time: DSN em runtime e source maps servidos pelo app

**Status:** Accepted

## Context

O plano original (registrado em `WORKSHOP.md` e `INFRASTRUCTURE.md`) era o
frontend ler o DSN de `VITE_SENTRY_DSN`. Duas coisas se chocam nesse caminho:

- O **Vite congela** qualquer variável `VITE_*` dentro do bundle JavaScript no
  momento do `npm run build` — que, aqui, acontece dentro do `Dockerfile`.
- O **Kamal entrega secrets ao container só em tempo de execução**
  (`env.secret`), depois que a imagem já está construída.

O participante do workshop criaria o secret, a pipeline ficaria **verde**, o app
funcionaria normalmente, e nenhum erro de navegador chegaria ao Sentry. Sem
mensagem de erro, sem log, sem pista. É a mesma classe de falha silenciosa do
[ADR-003](003-spa-em-producao-e-dev-mode.md): saudável por fora, quebrado por
dentro — e num workshop de 1h30 não há tempo para depurar isso ao vivo.

## Decision

O backend expõe **`GET /api/config`**, público e sem autenticação, devolvendo
`sentry_dsn`, `environment` e `release`. A SPA lê esse endpoint na inicialização
e só então chama `Sentry.init`; DSN vazio significa "não inicialize".

Um único secret `SENTRY_DSN` atende backend e navegador. `VITE_SENTRY_DSN`
deixa de existir.

Junto disso, `InitSentry` passa a receber `APP_ENV` (`local`, `preview`) em vez
de `BASE_URL` — o campo é o facet de ambiente do Sentry, não um endereço.

**Source maps pela mesma lógica.** O build do Vite emite os `.map`
(`build.sourcemap: true`), o handler de estáticos os serve publicamente e o
Sentry os busca pela URL do `sourceMappingURL`. Sem isso, toda Issue de
frontend chega minificada (`a.b is not a function` em
`index-4f2a.js:1:20481`) — inútil para entregar a um agente.

**Não usar o wizard do Sentry** (`npx @sentry/wizard`) neste projeto. Ele
configura o `@sentry/vite-plugin`, que exige um `SENTRY_AUTH_TOKEN` **no build
da imagem** — a mesma armadilha, com um secret a mais por participante. Ele
também grava o DSN no código-fonte (o que obrigaria cada fork a editar código,
já que cada participante tem o seu) e é interativo, abrindo o navegador para
login.

## Rationale

Expor o DSN é seguro **por projeto**: ele é uma credencial de escrita, desenhada
para viajar dentro do JavaScript servido ao navegador. Qualquer integração de
frontend com Sentry, inclusive a oficial, deixa o DSN visível no bundle.

Entregá-lo em runtime elimina a armadilha e ainda dá três ganhos práticos:
trocar o DSN não exige reconstruir a imagem; os dois lados compartilham o mesmo
`release`, então erros de servidor e de navegador se alinham na interface do
Sentry; e o endpoint vira o lugar natural para qualquer configuração pública
futura.

## Trade-offs

**Pros:**
- Impossível a pipeline "passar verde" com o Sentry silenciosamente desligado.
- Um secret em vez de dois (três, contando o auth token); trocar o DSN não
  exige rebuild.
- `environment` e `release` consistentes entre backend e frontend.
- Nada de novo no build: nem variável, nem token, nem etapa de upload.

**Cons:**
- A SPA faz uma requisição a mais no boot.
- Erros que aconteçam **antes** de `/api/config` responder não são capturados.
  Aceitável: o app depende do backend de qualquer forma.
- Foge da documentação oficial do Sentry para Vite, então merece o comentário
  que está em `backend/internal/handler/appconfig.go`.
- Os source maps expõem o código-fonte do frontend a quem abrir o app. Aceito:
  o repositório é público e os forks também.
- A leitura dos maps depende de **"Allow JavaScript source fetching"** na org
  do Sentry (ligado por padrão). Está na tabela de troubleshooting do README.
- Bloqueadores de anúncio barram `*.ingest.sentry.io` e derrubam só os eventos
  de navegador — sintoma confuso, também documentado no README.

## Alternatives Considered

- **`VITE_SENTRY_DSN` como build arg do Docker (via `builder.secrets` do
  Kamal):** descartado — funciona, mas grava o DSN dentro da imagem, obriga a
  reconstruir para trocá-lo, exige dois secrets e adiciona configuração de build
  justamente no ponto do workshop que precisa ser mais simples.
- **Injetar o DSN no `index.html` no momento de servir:** descartado — obrigaria
  o Go a reescrever HTML a cada requisição, quebrando o cache de arquivo
  estático, para ganhar apenas uma requisição.
- **Um projeto Sentry separado para o frontend:** descartado — dobra o setup
  manual do participante; a tag `environment` já separa o que precisa ser
  separado.
- **Subir os source maps com o `@sentry/vite-plugin`** (o caminho que o wizard
  configura): descartado — exige `SENTRY_AUTH_TOKEN` no build da imagem, mais
  `org` e `project` fixos no código, que cada fork teria de trocar. É a forma
  recomendada pelo Sentry em produção de verdade; para este workshop custa três
  configurações a mais para resolver um problema que servir os `.map` já
  resolve.
- **Tunelar os eventos por uma rota do próprio app** para escapar de
  bloqueadores: descartado por ora — exigiria o Go fazer proxy para o Sentry,
  com o custo de banda e um ponto novo de falha. Uma linha no troubleshooting
  resolve o caso do workshop. Reavaliar se muita gente tropeçar nisso.
