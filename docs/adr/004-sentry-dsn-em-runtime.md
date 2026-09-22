# 004 - DSN do Sentry entregue em runtime, não em build time

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
- Um secret em vez de dois; trocar o DSN não exige rebuild.
- `environment` e `release` consistentes entre backend e frontend.

**Cons:**
- A SPA faz uma requisição a mais no boot.
- Erros que aconteçam **antes** de `/api/config` responder não são capturados.
  Aceitável: o app depende do backend de qualquer forma.
- Foge da documentação oficial do Sentry para Vite, então merece o comentário
  que está em `backend/internal/handler/appconfig.go`.

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
