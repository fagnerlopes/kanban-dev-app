# PRD — Kanban Dev Flow

## Overview

Kanban Dev Flow é uma aplicação web de quadro kanban para equipes de
desenvolvimento. Ela permite visualizar o fluxo de trabalho em colunas
(Backlog, To Do, In Dev, Review, Done), criar tarefas, movê-las entre
colunas (drag-and-drop) e removê-las.

É uma aplicação de demonstração para um workshop de ChatOps: o foco é
mostrar um agente de IA (Hermes) interagindo com um projeto real de
software — forking, publicando, integrando com Sentry e corrigindo bugs.

## Target Users

- Desenvolvedores que acompanham o progresso de tarefas em um quadro.
- Participantes do workshop que interagem com o app para gerar eventos de
  erro no Sentry.

## Core Features

1. **Quadro kanban** — 5 colunas fixas: Backlog, To Do, In Dev, Review, Done.
2. **Criar tarefa** — título e descrição opcional, adicionada em uma coluna.
3. **Mover tarefa** — drag-and-drop entre colunas (e reordenação).
4. **Remover tarefa** — excluir uma tarefa do quadro.
5. **Tema claro/escuro** — toggle persistido em `localStorage`.
6. **Login mockado** — um único usuário de demonstração; autenticação
   simplificada (sem fluxo real de OAuth/magic link).

## User Flows

### Criar e mover uma tarefa
1. Usuário abre o app (já logado como usuário demo).
2. Clica em "Nova tarefa", escolhe a coluna, digita o título.
3. A tarefa aparece na coluna.
4. O usuário arrasta a tarefa para outra coluna.
5. O quadro reflete a mudança imediatamente.

### Alterar tema
1. Usuário clica no ícone de tema (sol/lua).
2. O tema alterna entre claro e escuro.
3. A escolha persiste entre sessões (localStorage).

## Non-Functional Requirements

- **Performance:** respostas da API < 200ms local.
- **Disponibilidade:** health check em `GET /up` para o orquestrador.
- **Observabilidade:** erros de runtime reportados ao Sentry (quando o DSN
  está configurado).
- **Segurança:** em produção, `DEV_MODE` nunca é definido; rotas de dev
  não existem.

## Out of Scope

- Autenticação real (OAuth, magic link, senha) — fora do escopo da demo.
- Múltiplos quadros / boards por usuário.
- Colaboração em tempo real (SSE) — o estado é sincronizado por refresh.
- Permissões / multi-tenancy.
- Busca, filtros, etiquetas, anexos.
