#!/usr/bin/env bash
#
# Aponta um dominio proprio para o app, ou volta para o nip.io.
#
#   ./scripts/set-domain.sh                         # pergunta o dominio
#   ./scripts/set-domain.sh kanbandev.exemplo.dev   # direto
#   ./scripts/set-domain.sh --reset                 # volta para o nip.io

set -euo pipefail

bold() { printf '\033[1m%s\033[0m\n' "$1"; }
ok()   { printf '  \033[32m/\033[0m %s\n' "$1"; }
warn() { printf '  \033[33m!\033[0m %s\n' "$1"; }
die()  { printf '\n\033[31mErro:\033[0m %s\n' "$1" >&2; exit 1; }

command -v gh >/dev/null 2>&1 || die "o GitHub CLI (gh) nao esta instalado."
gh auth status >/dev/null 2>&1 || die "voce nao esta logado no GitHub. Rode: gh auth login"
gh repo view >/dev/null 2>&1 || die "rode este script de dentro do seu fork ja clonado."

ROOT=$(git rev-parse --show-toplevel 2>/dev/null) || die "isto nao e um repositorio git."

redeploy() { exec make -C "$ROOT" --no-print-directory deploy; }

# Resolve pelo DNS publico: quem valida o dominio e o Let's Encrypt, a partir
# da internet. DNS corporativo ou VPN daria falso negativo.
resolve_ips() {
  local name="$1" found="" server out have_tool=0

  if command -v dig >/dev/null 2>&1; then
    have_tool=1
    for server in 1.1.1.1 8.8.8.8; do
      out=$(dig +short +time=3 +tries=1 A "$name" "@$server" 2>/dev/null || true)
      found="$found$out
"
    done
  fi

  if command -v getent >/dev/null 2>&1; then
    have_tool=1
    out=$(getent ahostsv4 "$name" 2>/dev/null | awk '{print $1}' || true)
    found="$found$out
"
  elif command -v python3 >/dev/null 2>&1; then
    have_tool=1
    out=$(python3 -c "import socket,sys
try: print(socket.gethostbyname(sys.argv[1]))
except Exception: pass" "$name" 2>/dev/null || true)
    found="$found$out
"
  fi

  [ "$have_tool" -eq 1 ] || return 1
  printf '%s\n' "$found" | grep -E '^[0-9]+(\.[0-9]+){3}$' | sort -u || true
  return 0
}

web_ip() {
  local run tmp ip
  run=$(gh run list --workflow "Deploy Preview" --status success --limit 1 \
        --json databaseId -q '.[0].databaseId' 2>/dev/null) || return 1
  [ -n "$run" ] || return 1
  tmp=$(mktemp -d)
  if gh run download "$run" --name provision-output --dir "$tmp" >/dev/null 2>&1; then
    ip=$(grep -o '"web_ip"[^,]*' "$tmp/provision-output.json" | head -1 | cut -d'"' -f4)
  fi
  rm -rf "$tmp"
  printf '%s' "${ip:-}"
}

# ------------------------------------------------------------------ reset ---
if [ "${1:-}" = "--reset" ]; then
  gh variable delete APP_DOMAIN >/dev/null 2>&1 \
    && ok "APP_DOMAIN removida" \
    || warn "APP_DOMAIN ja nao existia"
  echo
  bold "Republicando para voltar ao endereco nip.io..."
  redeploy
fi

# --------------------------------------------------------------- dominio ---
DOMAIN="${1:-}"
if [ -z "$DOMAIN" ]; then
  echo
  echo "  Informe o dominio que vai apontar para o app."
  echo "  Exemplo: kanbandev.suaempresa.dev"
  echo "  Para um dominio raiz, inclua o www separado por virgula:"
  echo "    exemplo.com.br,www.exemplo.com.br"
  echo
  printf '  Dominio (ENTER cancela): '
  read -r DOMAIN || true
  [ -z "$DOMAIN" ] && { echo "Cancelado, nada mudou."; exit 0; }
fi

# Normaliza: tira espacos, esquema e barra final que as pessoas costumam colar.
DOMAIN=$(printf '%s' "$DOMAIN" \
  | tr -d ' ' \
  | sed -e 's#^https\?://##' -e 's#/$##')

printf '%s' "$DOMAIN" | grep -Eq '^[A-Za-z0-9.-]+(,[A-Za-z0-9.-]+)*$' \
  || die "dominio invalido: $DOMAIN"

echo
bold "Conferindo o DNS"

IP=$(web_ip)
if [ -z "$IP" ]; then
  warn "nao achei o IP da VM (nenhum deploy bem-sucedido ainda?)."
  warn "publique uma vez com 'make deploy' antes de configurar o dominio."
  exit 1
fi
ok "VM web: $IP"

MISMATCH=0
OLDIFS=$IFS; IFS=','
for NAME in $DOMAIN; do
  IFS=$OLDIFS
  if ! RESOLVED=$(resolve_ips "$NAME"); then
    warn "sem ferramenta de DNS nesta maquina; pulando a conferencia"
    break
  fi
  if [ -z "$RESOLVED" ]; then
    warn "$NAME ainda nao resolve"
    MISMATCH=1
  elif printf '%s\n' "$RESOLVED" | grep -qx "$IP"; then
    ok "$NAME -> $IP"
  else
    warn "$NAME resolve para $(printf '%s' "$RESOLVED" | tr '\n' ' '), e nao para $IP"
    MISMATCH=1
  fi
  IFS=','
done
IFS=$OLDIFS

if [ "$MISMATCH" -eq 1 ]; then
  cat <<EOF

  O DNS ainda nao esta pronto. Crie no seu provedor de dominio:

      Tipo: A     Nome: $(printf '%s' "$DOMAIN" | cut -d, -f1)     Valor: $IP     TTL: 300

  (um registro para cada nome, se voce informou mais de um)

  A propagacao costuma levar de 1 a 30 minutos. O certificado TLS e emitido
  por desafio HTTP-01, entao o dominio so vai abrir depois que o DNS apontar
  para ca -- o deploy passa normalmente, ele so nao responde nesse nome.

  Seguir agora nao quebra nada: o endereco nip.io continua servindo o app.
  Quando o DNS propagar, rode este comando de novo.

EOF
  printf '  Continuar mesmo assim? [s/N]: '
  read -r GO || true
  case "${GO:-}" in
    s|S|y|Y) ;;
    *) echo "Cancelado, nada mudou."; exit 0 ;;
  esac
fi

echo
gh variable set APP_DOMAIN --body "$DOMAIN"
ok "APP_DOMAIN = $DOMAIN"
echo
bold "Republicando para o app passar a responder nesse endereco..."
redeploy
