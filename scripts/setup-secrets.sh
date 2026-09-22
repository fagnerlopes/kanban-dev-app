#!/usr/bin/env bash
#
# Cria, no SEU fork do GitHub, os secrets que a pipeline de deploy precisa.
#
#   CLOUDSTACK_API_KEY     voce cola (painel da Locaweb Cloud)
#   CLOUDSTACK_SECRET_KEY  voce cola (painel da Locaweb Cloud)
#   SSH_PRIVATE_KEY        gerado aqui
#   POSTGRES_PASSWORD      gerado aqui
#   SENTRY_DSN             voce cola (painel do Sentry) -- opcional
#
# E idempotente: secrets que ja existem sao mantidos. Use --force para
# recriar tudo do zero.
#
#   ./scripts/setup-secrets.sh
#   ./scripts/setup-secrets.sh --force

set -euo pipefail

FORCE=0
[ "${1:-}" = "--force" ] && FORCE=1

bold() { printf '\033[1m%s\033[0m\n' "$1"; }
ok()   { printf '  \033[32m/\033[0m %s\n' "$1"; }
skip() { printf '  \033[90m-\033[0m %s\n' "$1"; }
warn() { printf '  \033[33m!\033[0m %s\n' "$1"; }
die()  { printf '\n\033[31mErro:\033[0m %s\n' "$1" >&2; exit 1; }

# ---------------------------------------------------------------- prereqs ---
command -v gh >/dev/null 2>&1 || die "o GitHub CLI (gh) nao esta instalado. Veja https://cli.github.com/"
command -v ssh-keygen >/dev/null 2>&1 || die "ssh-keygen nao encontrado (instale o pacote openssh-client)."
gh auth status >/dev/null 2>&1 || die "voce nao esta logado no GitHub. Rode: gh auth login"

REPO=$(gh repo view --json nameWithOwner -q .nameWithOwner) \
  || die "rode este script de dentro do seu fork ja clonado."
REPO_NAME=${REPO##*/}
KEY_PATH="$HOME/.ssh/$REPO_NAME"

bold ""
bold "Configurando os secrets de $REPO"
echo

EXISTING=$(gh secret list --json name -q '.[].name' 2>/dev/null || true)
has_secret() {
  [ "$FORCE" -eq 1 ] && return 1
  printf '%s\n' "$EXISTING" | grep -qx "$1"
}

# ----------------------------------------------------------- 1. chave SSH ---
bold "1/4  Chave SSH (o Kamal usa para entrar nas VMs)"
if [ -f "$KEY_PATH" ]; then
  skip "chave ja existe em $KEY_PATH (mantida)"
else
  ssh-keygen -t ed25519 -f "$KEY_PATH" -N "" -C "$REPO_NAME-deploy" >/dev/null
  chmod 600 "$KEY_PATH"
  ok "chave gerada em $KEY_PATH"
fi
if has_secret SSH_PRIVATE_KEY; then
  skip "secret SSH_PRIVATE_KEY ja existe"
else
  gh secret set SSH_PRIVATE_KEY < "$KEY_PATH"
  ok "secret SSH_PRIVATE_KEY gravado"
fi
echo

# ---------------------------------------------------- 2. senha do Postgres ---
bold "2/4  Senha do Postgres"
if has_secret POSTGRES_PASSWORD; then
  skip "secret POSTGRES_PASSWORD ja existe"
else
  # Alfanumerica de proposito: a senha e interpolada dentro da DATABASE_URL,
  # entao caracteres como @ / # so criariam armadilhas.
  PG_PASS=$(LC_ALL=C tr -dc 'A-Za-z0-9' < /dev/urandom | head -c 40)
  printf '%s' "$PG_PASS" | gh secret set POSTGRES_PASSWORD
  unset PG_PASS
  ok "secret POSTGRES_PASSWORD gerado e gravado (nunca aparece na tela)"
fi
echo

# ------------------------------------------------- 3. credenciais da nuvem ---
bold "3/4  Credenciais da Locaweb Cloud"
if has_secret CLOUDSTACK_API_KEY && has_secret CLOUDSTACK_SECRET_KEY; then
  skip "CLOUDSTACK_API_KEY e CLOUDSTACK_SECRET_KEY ja existem"
else
  cat <<'EOF'
  Pegue as duas chaves em https://painel-cloud.locaweb.com.br/
    Contas > (sua conta) > Visualizar usuarios > (seu usuario)
  Elas carregam de forma assincrona: espere ate 60s. Se nao aparecerem,
  clique em "Gerar novas chaves" no canto superior direito.

EOF
  if [ ! -t 0 ]; then
    warn "sem terminal interativo; crie os dois secrets na mao em:"
    warn "$(gh repo view --json url -q .url)/settings/secrets/actions"
  else
    for NAME in CLOUDSTACK_API_KEY CLOUDSTACK_SECRET_KEY; do
      if has_secret "$NAME"; then
        skip "$NAME ja existe"
        continue
      fi
      VALUE=""
      while [ -z "$VALUE" ]; do
        printf '  Cole %s (fica invisivel): ' "$NAME"
        read -rs VALUE || true
        echo
        [ -z "$VALUE" ] && warn "valor vazio, tente de novo"
      done
      printf '%s' "$VALUE" | gh secret set "$NAME"
      unset VALUE
      ok "$NAME gravado"
    done
  fi
fi
echo

# --------------------------------------------------------------- 4. Sentry ---
bold "4/4  Sentry (opcional -- pode deixar para depois)"
if has_secret SENTRY_DSN; then
  skip "secret SENTRY_DSN ja existe"
elif [ ! -t 0 ]; then
  skip "sem terminal interativo; configure depois com: make sentry"
else
  echo "  O DSN esta em: Sentry > Settings > Projects > (seu projeto) > Client Keys (DSN)"
  printf '  Cole o SENTRY_DSN (ENTER pula): '
  read -r DSN || true
  if [ -n "${DSN:-}" ]; then
    printf '%s' "$DSN" | gh secret set SENTRY_DSN
    ok "secret SENTRY_DSN gravado"
  else
    skip "pulado -- o app sobe com o Sentry desligado; rode 'make sentry' quando tiver o DSN"
  fi
  unset DSN
fi
echo

bold "Pronto. Secrets configurados em $REPO:"
gh secret list
echo
bold "Proximo passo: publicar"
echo "  make deploy      # dispara o deploy e acompanha ate o fim"
echo
