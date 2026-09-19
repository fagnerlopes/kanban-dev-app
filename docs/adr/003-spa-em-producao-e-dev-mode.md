# 003 - Servir a SPA em produção e o escopo do DEV_MODE

**Status:** Accepted

## Context

O preview subia "saudável" — `GET /up` retornava 200 e `GET /api/board`
devolvia o quadro — mas **toda página HTML retornava 404**. O app estava no
ar e inutilizável, e o health check não acusava nada.

A causa: `config/deploy.preview.yml` definia `DEV_MODE: "1"` ("workshop
demo: enables mocked login + dev endpoints"). No servidor Go, `DEV_MODE`
significa *"o Vite está servindo o frontend"* — então o handler de arquivos
estáticos **não é registrado**. Em desenvolvimento isso é correto; em um
ambiente publicado, derruba a interface inteira.

O motivo de terem ligado a flag: o login chamava `POST /api/dev/login`, uma
rota que só existe com `DEV_MODE=1`. Sem a flag o login quebrava; com a
flag, a SPA sumia. Os dois caminhos estavam errados.

## Decision

- `DEV_MODE` é **exclusivamente local**. Nenhum arquivo `config/deploy*.yml`
  pode defini-lo — garantido por teste (`TestDeployConfigsNeverSetDevMode`).
- O login mockado passa a ser **inteiramente client-side** (`useAuth`): a
  sessão demo vive no `localStorage`, sem chamada de rede. Não depende mais
  de nenhuma rota de dev.
- `POST /api/dev/login` continua existindo apenas sob `DEV_MODE=1`, como
  exige o padrão Cofounder para testar rotas autenticadas localmente.
- O servidor de arquivos estáticos saiu do `main.go` para
  `handler.RegisterFrontend`, com testes cobrindo cada regra: `/` serve o
  `index.html`, assets saem com o `Content-Type` correto, rotas da SPA caem
  no shell, páginas pré-renderizadas não redirecionam, diretórios nunca são
  listados e caminhos `/api/`ausentes devolvem 404.

## Rationale

O health check `/up` não prova que a aplicação está utilizável — ele prova
que o processo está de pé. A regressão passou por três deploys "verdes".
Cobrir o serviço de arquivos estáticos com teste transforma uma falha
silenciosa e só visível em produção em uma falha na bateria de testes.

Manter a sessão demo no cliente também é mais honesto: não existe sessão no
servidor nem dado por usuário, então o login sempre foi um portão de
interface, nunca uma fronteira de segurança.

## Trade-offs

**Pros:**
- A interface funciona em qualquer ambiente publicado.
- `DEV_MODE` volta a ter um único significado, protegido por teste.
- A regressão específica não pode voltar sem quebrar o CI.

**Cons:**
- O login demo não exercita mais nenhuma rota do backend.
- A sessão demo é falsificável pelo usuário (aceitável: não há dado privado).

## Alternatives Considered

- **Registrar a SPA mesmo com `DEV_MODE=1`:** descartado — em
  desenvolvimento o Go passaria a servir um build obsoleto por cima do
  Vite, criando um bug pior e mais confuso.
- **Expor `/api/dev/login` sem a flag:** descartado — seria uma rota de
  desenvolvimento ativa em produção, exatamente o que o padrão Cofounder
  proíbe.
- **Remover a tela de login:** descartado — o PRD a prevê e ela faz parte
  do roteiro do workshop.
