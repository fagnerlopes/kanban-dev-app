# Atalhos do workshop. Tudo aqui roda contra o SEU fork no GitHub.
#
#   make setup    cria os secrets da Locaweb Cloud, SSH e Postgres
#   make deploy   publica na Locaweb Cloud e acompanha o deploy
#   make url      mostra o endereco do seu app no ar
#   make sentry   grava o DSN do Sentry e manda republicar
#   make domain   aponta um dominio proprio para o app
#   make status   mostra secrets, dominio e ultimos deploys

.DEFAULT_GOAL := help
.PHONY: help setup deploy url sentry domain domain-reset status

WORKFLOW := Deploy Preview

help:
	@echo ""
	@echo "  make setup    1. cria os secrets no seu fork (Locaweb Cloud, SSH, Postgres)"
	@echo "  make deploy   2. publica o app e acompanha ate terminar"
	@echo "  make url      3. mostra o endereco do app no ar"
	@echo "  make sentry   4. grava o DSN do Sentry (e republica)"
	@echo ""
	@echo "  make domain [d=meu.dominio.dev]   aponta um dominio proprio"
	@echo "  make domain-reset                 volta para o endereco nip.io"
	@echo "  make status                       secrets, dominio e deploys"
	@echo ""

setup:
	@bash scripts/setup-secrets.sh

deploy:
	@gh workflow run "$(WORKFLOW)" >/dev/null || { \
	  echo "Nao consegui disparar o workflow."; \
	  echo "Abra a aba Actions do seu fork e clique em \"I understand my workflows, go ahead and enable them\"."; \
	  exit 1; }
	@echo "Deploy disparado. Aguardando o GitHub registrar a execucao..."
	@sleep 8
	@RUN_ID=$$(gh run list --workflow "$(WORKFLOW)" --limit 1 --json databaseId -q '.[0].databaseId'); \
	  gh run list --workflow "$(WORKFLOW)" --limit 1 --json url -q '.[0].url' | sed 's/^/Acompanhe em: /'; \
	  gh run watch "$$RUN_ID" --exit-status || { \
	    echo ""; echo "O deploy falhou. Detalhe do erro:"; \
	    gh run view "$$RUN_ID" --log-failed | tail -40; exit 1; }
	@$(MAKE) --no-print-directory url

url:
	@DOMAIN=$$(gh variable list --json name,value -q '.[] | select(.name=="APP_DOMAIN") | .value' 2>/dev/null | cut -d, -f1); \
	  RUN_ID=$$(gh run list --workflow "$(WORKFLOW)" --status success --limit 1 --json databaseId -q '.[0].databaseId'); \
	  if [ -z "$$RUN_ID" ]; then echo "Nenhum deploy bem-sucedido ainda. Rode: make deploy"; exit 1; fi; \
	  TMP=$$(mktemp -d); \
	  gh run download "$$RUN_ID" --name provision-output --dir "$$TMP" >/dev/null 2>&1 || { \
	    echo "Nao encontrei o resultado do provisionamento nesse deploy."; rm -rf "$$TMP"; exit 1; }; \
	  IP=$$(grep -o '"web_ip"[^,]*' "$$TMP/provision-output.json" | head -1 | cut -d'"' -f4); \
	  rm -rf "$$TMP"; \
	  if [ -z "$$IP" ]; then echo "Nao consegui ler o IP da VM web."; exit 1; fi; \
	  echo ""; \
	  if [ -n "$$DOMAIN" ]; then \
	    echo "  Seu app esta no ar:  https://$$DOMAIN"; \
	    echo "  (endereco direto da VM:  https://$$IP.nip.io)"; \
	  else \
	    echo "  Seu app esta no ar:  https://$$IP.nip.io"; \
	  fi; \
	  echo ""

domain:
	@bash scripts/set-domain.sh $(d)

domain-reset:
	@bash scripts/set-domain.sh --reset

sentry:
	@printf 'Cole o SENTRY_DSN (Sentry > Settings > Projects > seu projeto > Client Keys): '; \
	  read -r DSN; \
	  if [ -z "$$DSN" ]; then echo "Nada colado, nada mudou."; exit 1; fi; \
	  printf '%s' "$$DSN" | gh secret set SENTRY_DSN
	@echo "Secret gravado. Republicando para o app passar a enviar erros ao Sentry..."
	@$(MAKE) --no-print-directory deploy

status:
	@echo ""; echo "Secrets:"; gh secret list
	@echo ""; echo "Dominio:"; \
	  DOMAIN=$$(gh variable list --json name,value -q '.[] | select(.name=="APP_DOMAIN") | .value' 2>/dev/null); \
	  if [ -n "$$DOMAIN" ]; then echo "  APP_DOMAIN = $$DOMAIN"; \
	  else echo "  nenhum (o app responde no endereco nip.io da VM)"; fi
	@echo ""; echo "Ultimos deploys:"; gh run list --workflow "$(WORKFLOW)" --limit 5
	@echo ""
