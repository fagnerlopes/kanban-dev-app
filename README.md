# Kanban Dev Flow — workshop de ChatOps com o Hermes Agent

Um quadro Kanban de verdade, publicado na sua própria conta da **Locaweb Cloud**,
monitorado pelo **Sentry** — e consertado pelo **Hermes Agent** através do
**Telegram**.

Você não vai escrever código neste workshop. Você vai **forkar**, **publicar**,
**ligar o Sentry** e, quando os erros aparecerem, **pedir pelo Telegram** para o
Hermes investigar e corrigir. Ele abre o Pull Request e publica a correção.

O passo a passo abaixo leva cerca de **30 minutos** até o app estar no ar com o
Sentry ligado.

---

## O que você vai colocar no ar

```
        Você  ──►  https://<ip>.nip.io          (TLS automático, Let's Encrypt)
                          │
                          ▼
        ┌─────────────────────────────┐        ┌──────────────────────────┐
        │  VM 1 · web                 │        │  VM 2 · db               │
        │  Go + SPA React, porta 80   │───────►│  PostgreSQL 17           │
        │  health check em /up        │        │  dados em /data/pgdata   │
        └─────────────────────────────┘        └──────────────────────────┘
                          │
                          └──────────────►  Sentry (SaaS)  ── erros do servidor
                                                           └─ erros do navegador
```

**Duas VMs**, provisionadas automaticamente pela pipeline do GitHub Actions. Você
não abre o painel da nuvem para criar nada — só para pegar duas chaves de API.

| Componente | Tecnologia |
|---|---|
| Backend | Go (stdlib `net/http`, pgx, sqlc) |
| Frontend | React + React Router (SPA) + Tailwind + shadcn |
| Banco | PostgreSQL 17 (`supabase/postgres`) |
| Deploy | GitHub Actions + [Kamal](https://kamal-deploy.org/) na Locaweb Cloud |
| Erros | Sentry (plano gratuito) |

---

## Antes de começar

Você precisa de:

- Uma conta no **GitHub** com o [`gh` CLI](https://cli.github.com/) instalado e
  autenticado (`gh auth login`).
- Uma conta na **[Locaweb Cloud](https://www.locaweb.com.br/locaweb-cloud/)**
  (se não tiver, clique em **Contratar**).
- O **Hermes Agent** rodando e conectado ao seu Telegram (feito na primeira
  parte do workshop).
- `git`, `make` e `ssh-keygen` — já vêm em Linux e macOS; no Windows, use o
  **Git Bash** ou o **WSL**.

Uma conta gratuita no **Sentry** será criada no passo 5 — não precisa ter antes.

---

## Passo 1 — Faça o fork

Abra <https://github.com/fagnerlopes/kanban-dev-app> e clique em **Fork**
(canto superior direito) → **Create fork**.

Agora clone o **seu** fork e entre na pasta:

```bash
gh repo clone SEU-USUARIO/kanban-dev-app
cd kanban-dev-app
```

> Troque `SEU-USUARIO` pelo seu usuário do GitHub.

---

## Passo 2 — Ligue o GitHub Actions no fork

O GitHub **desativa** as automações em forks por segurança. Ligue uma vez:

1. Abra a aba **Actions** do seu fork.
2. Clique no botão verde
   **"I understand my workflows, go ahead and enable them"**.

Sem isso, nada publica — e o GitHub não avisa o porquê.

---

## Passo 3 — Pegue as chaves da Locaweb Cloud

1. Acesse <https://painel-cloud.locaweb.com.br/>.
2. Vá em **Contas** → *(sua conta)* → **Visualizar usuários** → *(seu usuário)*.
3. **Espere até 60 segundos.** As chaves carregam de forma assíncrona e
   aparecem abaixo da data em **"Criado"**.
4. Se não aparecerem, clique no ícone **"Gerar novas chaves"** no canto
   superior direito e espere de novo.

Deixe a página aberta — você vai copiar **Chave da API** e **Chave secreta** no
próximo passo.

---

## Passo 4 — Crie os secrets

Um comando resolve tudo:

```bash
make setup
```

O script vai:

- gerar uma **chave SSH** exclusiva deste projeto (`~/.ssh/kanban-dev-app`);
- sortear uma **senha do Postgres** (ela nunca aparece na tela);
- pedir que você **cole as duas chaves** da Locaweb Cloud — a digitação fica
  invisível, como numa senha;
- perguntar pelo **DSN do Sentry**, que você ainda não tem. **Aperte ENTER e
  pule** — voltamos nele no passo 6.

No fim ele lista o que foi criado:

| Secret | De onde vem |
|---|---|
| `CLOUDSTACK_API_KEY` | painel da Locaweb Cloud (você cola) |
| `CLOUDSTACK_SECRET_KEY` | painel da Locaweb Cloud (você cola) |
| `SSH_PRIVATE_KEY` | gerado pelo `make setup` |
| `POSTGRES_PASSWORD` | gerado pelo `make setup` |
| `SENTRY_DSN` | painel do Sentry (passo 6) — opcional |

O comando é seguro de repetir: ele **não sobrescreve** nada que já exista. Para
recriar tudo do zero, use `make setup --force`.

<details>
<summary>Prefere fazer na mão, pelo site do GitHub?</summary>

Vá em **Settings → Secrets and variables → Actions → New repository secret** e
crie os quatro secrets da tabela acima. Para o `SSH_PRIVATE_KEY`, gere a chave
antes e cole o conteúdo do arquivo **privado** (o sem `.pub`):

```bash
ssh-keygen -t ed25519 -f ~/.ssh/kanban-dev-app -N "" -C "kanban-dev-app-deploy"
cat ~/.ssh/kanban-dev-app
```
</details>

---

## Passo 5 — Publique

```bash
make deploy
```

O comando dispara a pipeline e acompanha ao vivo. Nos bastidores ela:

1. provisiona as **duas VMs**, rede, disco, IP público e firewall;
2. constrói a imagem Docker (React + Go numa imagem só);
3. sobe o Postgres, roda as migrations e publica o app;
4. emite o certificado TLS via Let's Encrypt.

A **primeira vez leva de 5 a 10 minutos** (é quando a infraestrutura nasce). As
publicações seguintes levam cerca de 2 minutos.

Terminou, ele imprime o endereço:

```
  Seu app esta no ar:  https://191.252.226.176.nip.io
```

Abra no navegador e faça o **login demo**. A qualquer momento você recupera o
endereço com:

```bash
make url
```

> **O que é `nip.io`?** Um serviço de DNS que transforma qualquer IP num
> domínio. `191.252.226.176.nip.io` aponta para `191.252.226.176`. É o que
> permite ter HTTPS sem você comprar um domínio.

---

## Passo 6 — Integre o Sentry

O Sentry é quem vai **capturar os erros automaticamente** e virar a fonte das
conversas com o Hermes no Telegram.

### 6.1 — Crie a conta e o projeto

1. Crie a conta gratuita em <https://sentry.io/signup/>.
2. Crie um projeto:
   - **Platform:** escolha **Go**
   - **Alert frequency:** *Alert me on every new issue*
   - **Project name:** `kanban-dev-app`
3. Pule a tela de instruções de instalação — o código já está pronto.
4. Copie o **DSN** em
   **Settings → Projects → kanban-dev-app → Client Keys (DSN)**.
   Ele se parece com `https://abc123@o456.ingest.us.sentry.io/789`.

> **Um projeto só, para servidor e navegador.** Este app envia os dois tipos de
> erro para o mesmo DSN. Você filtra depois pela tag `environment` dentro do
> Sentry.

### 6.2 — Grave o DSN e republique

```bash
make sentry
```

Cole o DSN quando pedir. O comando grava o secret **e já republica** — porque o
app só passa a enxergar o DSN depois de uma nova publicação.

Pronto: **o backend já está reportando erros.** Não precisa mexer em código.

### 6.3 — Peça ao Hermes para ligar o Sentry no frontend

Esta parte é de propósito uma tarefa para o agente — é a demonstração de que o
Hermes constrói uma feature nova, não só conserta bug. Mande pelo **Telegram**:

> Integre o Sentry no frontend do app. **Não use o wizard do Sentry** e não leia
> o DSN de uma variável `VITE_*`: o backend já expõe `GET /api/config`, que
> devolve `sentry_dsn`, `environment` e `release`. Instale o `@sentry/react`,
> inicialize a partir dessa resposta, não inicialize quando o DSN vier vazio, e
> reporte também os erros que caem no `ErrorBoundary` do `root.tsx`. Os source
> maps já estão configurados — não mexa neles. Rode os testes, abra o PR e
> publique.

Quando ele terminar, confira em <https://sentry.io> que o projeto está
recebendo eventos.

<details>
<summary>Por que não usar o wizard do Sentry (<code>npx @sentry/wizard</code>)?</summary>

O wizard é ótimo num projeto pessoal, e errado aqui, por três motivos:

1. **Ele exige um `SENTRY_AUTH_TOKEN`** para subir os source maps — e esse token
   precisaria estar disponível **durante o build da imagem Docker**. É
   exatamente a mesma armadilha do `VITE_SENTRY_DSN`, só que com um secret a
   mais por participante. Aqui os source maps são resolvidos sem token nenhum
   (veja abaixo).
2. **Ele grava o DSN direto no código-fonte.** Num projeto só seu, tudo bem —
   o DSN é público mesmo. Mas cada participante tem o **seu** DSN: hardcoded,
   todo fork precisaria editar código antes de publicar.
3. **Ele é interativo** — abre o navegador para login no Sentry. Não funciona
   num fluxo por Telegram.

</details>

### 6.4 — Erros legíveis: os source maps

Isto **já está pronto**, não é um passo seu — mas vale saber por que funciona.

O JavaScript que roda no navegador é minificado. Sem ajuda, uma Issue do Sentry
diria `a.b is not a function` em `index-4f2a.js:1:20481` — inútil para pedir ao
Hermes "corrige isso". O build gera os **source maps** junto dos bundles, o
servidor Go os entrega publicamente, e o Sentry os busca sozinho pela URL. O
resultado é a Issue apontando o arquivo `.tsx`, a linha e o trecho de código
original.

Nenhum token, nenhum passo de upload, nenhum secret a mais. O preço é que o
código-fonte do frontend fica legível a partir do app publicado — o que aqui não
custa nada, já que o repositório é público.

> Se as stack traces continuarem minificadas, confira em **Sentry → Settings →
> Security & Privacy** se **"Allow JavaScript source fetching"** está ligado
> (vem ligado por padrão).

### ⚠️ Desligue o bloqueador de anúncios

**uBlock Origin, AdBlock, Brave Shields e o DNS da sua empresa bloqueiam
`*.ingest.sentry.io`.** Com qualquer um deles ativo, os erros de **navegador**
simplesmente não saem da sua máquina: o Sentry fica vazio, sem nenhum aviso, e
parece que a integração falhou.

Os erros de **servidor** continuam chegando normalmente — eles saem da VM, não
do seu navegador. É justamente isso que torna o sintoma confuso.

Antes do passo 7, desative o bloqueador para o domínio do seu app
(`https://<ip>.nip.io`) ou abra o app numa janela anônima sem extensões.

<details>
<summary>Por que não usar <code>VITE_SENTRY_DSN</code>?</summary>

O Vite **congela** as variáveis `VITE_*` dentro do JavaScript no momento em que
a imagem Docker é construída. O Kamal, por sua vez, só entrega os secrets ao
container **depois**, na hora de executar. O resultado seria o pior tipo de
falha: o deploy fica **verde**, o app funciona, e nenhum erro de navegador
chega ao Sentry — sem nenhuma mensagem explicando o porquê.

Por isso o DSN é servido em tempo de execução pelo backend, em `/api/config`.
Um único secret `SENTRY_DSN` atende os dois lados, e trocar o DSN não exige
reconstruir a imagem. Veja
[`docs/adr/004-sentry-dsn-em-runtime.md`](docs/adr/004-sentry-dsn-em-runtime.md).

E sim: expor o DSN publicamente é seguro. Ele é uma credencial de **escrita**,
projetada para viajar dentro do JavaScript do navegador.
</details>

---

## Opcional — Use o seu próprio domínio

Funciona perfeitamente sem isso: o endereço `nip.io` já tem HTTPS válido. Mas se
você tem um domínio e quer usá-lo, são dois comandos.

**A ordem importa.** O IP da sua VM só existe depois do primeiro deploy, e o
certificado é emitido por **desafio HTTP-01** — o Let's Encrypt acessa o seu
domínio para provar que ele é seu. Ou seja: o domínio só passa a abrir depois
que o DNS estiver apontando para a VM.

> **O endereço `nip.io` nunca para de funcionar.** Ele continua atendendo mesmo
> com um domínio configurado. Então errar o domínio ou configurá-lo antes do DNS
> propagar não derruba o seu app — o domínio só não abre ainda.

**1.** Descubra o IP da sua VM:

```bash
make url
```

**2.** No seu provedor de domínio, crie um registro **A**:

```
Tipo: A     Nome: kanbandev     Valor: <o IP do passo 1>     TTL: 300
```

**3.** Espere a propagação (costuma levar de 1 a 30 minutos) e rode:

```bash
make domain
```

Ele pergunta o domínio, **confere se o DNS já resolve para a sua VM** e só então
grava a configuração e republica. Se o DNS ainda não estiver pronto, ele avisa e
mostra exatamente o registro que falta criar.

A conferência consulta DNS público (1.1.1.1 e 8.8.8.8), não o resolvedor da sua
máquina — quem valida o domínio é o Let's Encrypt, a partir da internet. Se você
está em rede corporativa ou VPN, o DNS interno pode dizer que o domínio não
existe mesmo estando tudo certo.

Também dá para passar direto:

```bash
make domain d=kanbandev.suaempresa.dev
```

**Para voltar ao endereço `nip.io`:**

```bash
make domain-reset
```

<details>
<summary>Domínio raiz (<code>exemplo.com.br</code>, sem subdomínio)</summary>

Informe o domínio raiz **e** o `www`, separados por vírgula:

```bash
make domain d=exemplo.com.br,www.exemplo.com.br
```

Crie **dois** registros A, os dois apontando para o mesmo IP:

```
Tipo: A     Nome: @       Valor: <IP>     TTL: 300
Tipo: A     Nome: www     Valor: <IP>     TTL: 300
```

O primeiro da lista é o canônico — é ele que vira o endereço oficial do app. O
`www` também responde, mas não redireciona para o principal. Como a sessão demo
fica no navegador, entrar por um e depois pelo outro cria duas sessões
separadas. Para este workshop não atrapalha; num app de verdade, valeria um
redirecionamento.

</details>

<details>
<summary>Como isso funciona por dentro</summary>

`make domain` grava uma **variável de repositório** chamada `APP_DOMAIN`
(Settings → Secrets and variables → Actions → aba **Variables**). Não é um
secret: um domínio é público.

O `config/deploy.preview.yml` lê essa variável na hora do deploy e roteia **o
`nip.io` sempre, mais o seu domínio quando houver**. É por isso que um fork
recém-criado publica sem nenhuma configuração de DNS, e por isso que configurar
um domínio errado não tira o app do ar. Nada de editar arquivo e comitar: você
troca o domínio de um fork sem tocar no código.

Detalhe: `APP_DOMAIN` vale para o ambiente *preview*. Se um dia você criar um
ambiente de produção, ele usa o nome com sufixo (`APP_DOMAIN_PRODUCTION`),
mesma convenção dos secrets.

</details>

---

## Passo 7 — ChatOps: conserte os bugs pelo Telegram

A partir daqui o roteiro é conduzido ao vivo. O fluxo que você vai exercitar:

| # | O que você faz | O que o Hermes faz |
|---|---|---|
| 1 | Usa o app e encontra um erro | — |
| 2 | Abre o Sentry e vê a Issue | — |
| 3 | Pergunta no Telegram: *"o que houve nessa issue do Sentry?"* | Explica a causa em português claro |
| 4 | Pede: *"corrige isso"* | Investiga, corrige, roda os testes e manda o **link do PR** |
| 5 | Revisa e pede: *"publica"* | Faz o merge e dispara o deploy |
| 6 | Recarrega o app | — |

Mensagens que funcionam bem:

- *"Meu app está devolvendo 500. Descobre o motivo e corrige."*
- *"Tem uma Issue nova no Sentry. Me explica o que aconteceu."*
- *"Corrige o erro que acontece quando eu movo uma task de coluna."*
- *"Publica a correção e me avisa quando estiver no ar."*

---

## Comandos disponíveis

```
make setup          cria os secrets no seu fork (Locaweb Cloud, SSH, Postgres)
make deploy         publica o app e acompanha até terminar
make url            mostra o endereço do app no ar
make sentry         grava o DSN do Sentry e republica
make domain         aponta um domínio próprio (confere o DNS antes)
make domain-reset   volta para o endereço nip.io
make status         mostra secrets, domínio e últimos deploys
```

---

## Quando algo dá errado

| Sintoma | Causa provável | O que fazer |
|---|---|---|
| `make deploy` diz que não conseguiu disparar o workflow | Actions desativado no fork | Refaça o **passo 2** |
| A pipeline falha no primeiro job (`infra`) | Chaves da Locaweb Cloud erradas ou ausentes | `make status` para conferir; `make setup --force` para recriar |
| A pipeline falha logo no início com erro de credencial | Chave SSH gravada errada (faltou uma linha ao copiar) | `make setup --force` |
| O site não abre, mas `make url` mostra um endereço | O certificado TLS ainda está sendo emitido | Espere 1–2 minutos e recarregue |
| O app abre, mas o Sentry não recebe nada | O DSN foi gravado sem republicar | `make deploy` |
| Erros do **servidor** chegam ao Sentry, os do **navegador** não | Bloqueador de anúncios barrando `*.ingest.sentry.io` | Desative o bloqueador para o domínio do app, ou use uma janela anônima |
| Nenhum erro de navegador chega, sem bloqueador ativo | O frontend ainda não foi integrado | Faça o **passo 6.3** |
| A Issue mostra `a.b is not a function` em `index-4f2a.js:1:20481` | Source maps não estão sendo lidos | **Sentry → Settings → Security & Privacy → Allow JavaScript source fetching** |
| Configurei o domínio, o deploy passou, mas o domínio não abre | O DNS não aponta para a VM (ou ainda não propagou) — o deploy passa mesmo assim | `make url` para ver o IP, confira o registro A, espere propagar. Enquanto isso o endereço `nip.io` continua funcionando |
| Configurei o domínio e quero desfazer | — | `make domain-reset` |
| Quero apagar tudo da nuvem | — | Aba **Actions** → **Teardown Preview** → **Run workflow** |

Travou em algo que não está na tabela? **Pergunte ao Hermes pelo Telegram** —
ele tem acesso aos logs da pipeline e às VMs. É exatamente para isso que ele
está aqui.

---

## Rodar na sua máquina (opcional)

Não é necessário para o workshop, mas funciona:

```bash
# 1. Banco
podman run -d --name kanbandev-app-db -e POSTGRES_PASSWORD=postgres \
  -p 5432:5432 docker.io/supabase/postgres:17.6.1.171

# 2. Variáveis
cp .env.example .env     # ajuste DATABASE_URL se mudar a porta

# 3. Backend (terminal 1)
set -a && . .env && set +a
cd backend && mise x -- go run ./cmd/server

# 4. Frontend (terminal 2)
cd frontend && mise x -- npm install && mise x -- npm run dev
```

Abra <http://localhost:5173>. Testes:

```bash
cd backend  && mise x -- go test ./...   # Go
cd frontend && mise x -- npm test        # React
```

---

## Mapa do projeto

```
backend/          API em Go — handlers, migrations, queries (sqlc)
frontend/         SPA em React — rotas, componentes, hooks
config/           configuração do Kamal (deploy)
.kamal/           mapeamento dos secrets para o Kamal
.github/workflows/ pipelines de deploy e teardown
scripts/          scripts de apoio (setup dos secrets)
docs/             PRD, tarefas, infraestrutura e decisões (ADRs)
Dockerfile        imagem única: React + Go
```

Documentação mais funda em [`docs/`](docs/):
[PRD](docs/PRD.md) ·
[Tarefas](docs/TASKS.md) ·
[Infraestrutura](docs/INFRASTRUCTURE.md) ·
[Roteiro do workshop](docs/WORKSHOP.md) ·
[Decisões técnicas](docs/adr/)
