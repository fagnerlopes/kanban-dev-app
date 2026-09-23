# Kanban Dev Flow — workshop de ChatOps com o Hermes Agent

Um quadro Kanban publicado na sua conta da **Locaweb Cloud**, monitorado pelo
**Sentry** e consertado pelo **Hermes Agent** pelo **Telegram**.

Você não escreve código aqui. Você forka, publica, liga o Sentry e, quando os
erros aparecem, pede a correção pelo Telegram — o Hermes abre o PR e publica.

Do fork ao app no ar com Sentry: **cerca de 30 minutos.**

---

## O que você vai colocar no ar

```
   Você ──► https://<ip>.nip.io          (HTTPS automático)
                    │
        ┌───────────┴───────────┐        ┌──────────────────────┐
        │  VM 1 · web           │───────►│  VM 2 · db           │
        │  Go + React, porta 80 │        │  PostgreSQL 17       │
        └───────────────────────┘        └──────────────────────┘
                    │
                    └──────────► Sentry (erros do servidor e do navegador)
```

**Duas VMs**, provisionadas pela pipeline. Você não cria nada no painel da
nuvem — só pega duas chaves de API.

---

## Antes de começar

- Conta no **GitHub**, com o [`gh` CLI](https://cli.github.com/) instalado e
  autenticado (`gh auth login`).
- Conta na **[Locaweb Cloud](https://www.locaweb.com.br/locaweb-cloud/)**.
- **Hermes Agent** rodando e conectado ao seu Telegram.
- `git`, `make` e `ssh-keygen` — no Windows, use **Git Bash** ou **WSL**.

A conta do Sentry você cria no passo 6.

---

## 1. Fork e clone

Abra <https://github.com/fagnerlopes/kanban-dev-app>, clique em **Fork** →
**Create fork**. Depois:

```bash
gh repo clone SEU-USUARIO/kanban-dev-app
cd kanban-dev-app
```

## 2. Ligue o GitHub Actions

O GitHub **desativa** as automações em forks. Abra a aba **Actions** do seu fork
e clique em **"I understand my workflows, go ahead and enable them"**.

Sem isso nada publica — e o GitHub não avisa o porquê.

## 3. Pegue as chaves da Locaweb Cloud

1. Acesse <https://painel-cloud.locaweb.com.br/>.
2. **Contas** → *(sua conta)* → **Visualizar usuários** → *(seu usuário)*.
3. **Espere até 60 segundos** — as chaves carregam de forma assíncrona.
4. Se não aparecerem, clique em **"Gerar novas chaves"** no canto superior
   direito.

Deixe a página aberta.

## 4. Crie os secrets

```bash
make setup
```

O comando gera a chave SSH e a senha do Postgres, e pede que você **cole as duas
chaves** da Locaweb Cloud (a digitação fica invisível). Quando perguntar pelo
DSN do Sentry, **aperte ENTER e pule** — voltamos nele no passo 6.

É seguro repetir: não sobrescreve nada que já exista.

## 5. Publique

```bash
make deploy
```

A pipeline provisiona as duas VMs, constrói a imagem, roda as migrations e
emite o certificado. **A primeira vez leva de 5 a 10 minutos**; as seguintes,
cerca de 2.

No fim ele imprime o endereço. A qualquer momento, `make url` mostra de novo.

> `nip.io` é um DNS que transforma qualquer IP num domínio — é o que permite ter
> HTTPS sem você comprar um domínio.

## 6. Ligue o Sentry

**6.1** Crie a conta em <https://sentry.io/signup/> e um projeto com a
plataforma **Go**, nome `kanban-dev-app`. Pule as instruções de instalação: o
código já está pronto.

**6.2** Copie o **DSN** em **Settings → Projects → kanban-dev-app → Client Keys
(DSN)** e rode:

```bash
make sentry
```

Ele grava o secret **e já republica** — o app só enxerga o DSN depois de uma
nova publicação.

Pronto, o backend já reporta erros, logs, traces e métricas.

**6.3** Agora peça ao Hermes, pelo **Telegram**, para ligar o Sentry no
frontend:

> Integre o Sentry no frontend do app. **Não use o wizard do Sentry** e não leia
> o DSN de uma variável `VITE_*`: o backend já expõe `GET /api/config`, que
> devolve `sentry_dsn`, `environment` e `release`. Instale o `@sentry/react`,
> inicialize a partir dessa resposta, não inicialize quando o DSN vier vazio, e
> reporte também os erros que caem no `ErrorBoundary` do `root.tsx`. Os source
> maps já estão configurados — não mexa neles. Rode os testes, abra o PR e
> publique.

> **Desligue o bloqueador de anúncios.** uBlock Origin, AdBlock e Brave Shields
> bloqueiam `*.ingest.sentry.io`: os erros de **navegador** somem e o Sentry
> parece quebrado. Os de servidor continuam chegando, o que torna o sintoma
> confuso.

## 7. ChatOps: conserte os bugs pelo Telegram

| # | Você | O Hermes |
|---|---|---|
| 1 | Usa o app e encontra um erro | — |
| 2 | Abre o Sentry e vê a Issue | — |
| 3 | *"o que houve nessa issue do Sentry?"* | Explica a causa em português |
| 4 | *"corrige isso"* | Corrige, roda os testes e manda o **link do PR** |
| 5 | *"publica"* | Faz o merge e dispara o deploy |

Outras mensagens que funcionam bem:

- *"Meu app está devolvendo 500. Descobre o motivo e corrige."*
- *"Corrige o erro que acontece quando eu movo uma task de coluna."*
- *"Publica a correção e me avisa quando estiver no ar."*

## 8. Monitoramento com alerta no Telegram

A API expõe um health check em **`/api/health`**:

```bash
curl https://SEU-ENDERECO/api/health
```

```json
{
  "status": "ok",
  "checks": { "database": "ok" },
  "environment": "preview",
  "release": "kanban-dev-app",
  "uptime_seconds": 3847
}
```

Ele devolve **200** quando o app e o banco respondem, e **503** quando o banco
está fora. Peça ao Hermes, pelo **Telegram**:

> Crie um cronjob que consulte `https://SEU-ENDERECO/api/health` a cada minuto
> e me avise aqui no Telegram sempre que a resposta não for 200. Só alerte —
> não tente corrigir nada.

Depois desligue a VM web no [painel da Locaweb
Cloud](https://painel-cloud.locaweb.com.br/) e espere o alerta chegar. Para
voltar, ligue a VM de novo.

> **Por que não usar `/up`?** O `/up` é a sonda que o kamal-proxy usa para
> decidir se manda tráfego para o app — ele só diz que o processo está vivo, de
> propósito. O `/api/health` é o check profundo, que também consulta o banco.

---

## Domínio próprio (opcional)

O endereço `nip.io` já tem HTTPS. Se quiser usar seu domínio:

```bash
make url                 # veja o IP da VM
# crie um registro A apontando para esse IP, TTL 300
make domain              # confere o DNS e republica
```

`make domain` verifica se o DNS já resolve para a sua VM antes de publicar, e
mostra o registro exato que falta criar. Para um domínio raiz, informe o `www`
junto: `make domain d=exemplo.com.br,www.exemplo.com.br`.

O `nip.io` **nunca para de funcionar**, mesmo com domínio configurado — então
errar aqui não derruba o app. Para desfazer: `make domain-reset`.

---

## Comandos

```
make setup          cria os secrets no seu fork
make deploy         publica e acompanha até terminar
make url            mostra o endereço do app
make sentry         grava o DSN do Sentry e republica
make domain         aponta um domínio próprio
make domain-reset   volta para o endereço nip.io
make status         secrets, domínio e últimos deploys
```

---

## Quando algo dá errado

| Sintoma | O que fazer |
|---|---|
| `make deploy` não consegue disparar o workflow | Refaça o **passo 2** |
| A pipeline falha no primeiro job (`infra`) | `make status` para conferir os secrets; `make setup --force` para recriar |
| O site não abre logo após o deploy | O certificado ainda está sendo emitido: espere 1–2 minutos |
| O Sentry não recebe nada | O DSN foi gravado sem republicar: `make deploy` |
| Erros de servidor chegam, os de navegador não | Bloqueador de anúncios — veja o aviso no passo 6 |
| Nenhum erro de navegador, sem bloqueador | O frontend ainda não foi integrado: passo **6.3** |
| Configurei o domínio e ele não abre | O DNS não propagou; o `nip.io` segue funcionando |
| Quero apagar tudo da nuvem | Aba **Actions** → **Teardown Preview** → **Run workflow** |

Travou em outra coisa? **Pergunte ao Hermes pelo Telegram** — ele tem acesso aos
logs da pipeline e às VMs.

---

Documentação técnica em [`docs/`](docs/):
[PRD](docs/PRD.md) ·
[Infraestrutura](docs/INFRASTRUCTURE.md) ·
[Decisões](docs/adr/)
