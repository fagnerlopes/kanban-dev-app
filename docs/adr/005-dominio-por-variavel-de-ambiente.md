# 005 - Domínio personalizado por variável de repositório

**Status:** Accepted

## Context

O caminho padrão para apontar um domínio neste tipo de deploy é **editar
`config/deploy.<env>.yml`**, trocando `proxy.host` pelo domínio, comitar e
republicar.

Isso funciona num projeto com um dono. Aqui não: este repositório é um molde que
dezenas de participantes vão forkar. Se o domínio vive no arquivo, cada um
precisa editar código antes de publicar — e, pior, um domínio esquecido no
arquivo original faria **todo fork** pedir ao Let's Encrypt um certificado para
um nome que o participante não controla. O deploy falha na emissão, e a
mensagem de erro não diz que a causa é o DNS.

Há ainda uma ordem que não é óbvia: o IP da VM só existe **depois** do primeiro
deploy, e o certificado é emitido por desafio HTTP-01, que exige o domínio já
resolvendo para a VM **antes** do deploy. Quem tenta configurar na ordem
intuitiva (domínio primeiro) bate num deploy vermelho sem explicação.

## Decision

O hostname público vem de `APP_DOMAIN`, uma **variável de repositório** do
GitHub (não um secret — domínio é informação pública), lida pelo workflow e
resolvida em ERB no `config/deploy.preview.yml`:

- vazia → `<web_ip>.nip.io`, de modo que um fork publica sem nenhuma
  configuração de DNS;
- preenchida → o domínio, e `BASE_URL` acompanha;
- aceita lista separada por vírgula (`exemplo.com.br,www.exemplo.com.br`), com
  o primeiro item como canônico.

`make domain` encapsula o procedimento: normaliza o que foi colado (tira
`https://` e barra final), lê o IP da VM do último deploy, **confere se o DNS já
resolve para ele** e, quando não resolve, imprime o registro A exato que falta —
pedindo confirmação antes de seguir. `make domain-reset` volta ao `nip.io`.

`TestDeployConfigsKeepDomainConfigurable` garante que os configs continuem
lendo `APP_DOMAIN` e mantendo o fallback `nip.io`.

## Rationale

A escolha real não é "variável ou arquivo", é **onde mora o que difere entre
forks**. Tudo o mais que difere (chaves da nuvem, senha do banco, DSN) já vive
fora do código. O domínio é da mesma natureza, e tratá-lo como código obrigaria
cada participante a fazer um commit para algo que não é mudança de software.

A conferência de DNS existe porque o modo de falhar é péssimo: o deploy quebra
num passo que fala de certificado, não de DNS, e o participante não tem como
ligar uma coisa à outra no meio do workshop. Verificar antes transforma isso
numa instrução acionável.

## Trade-offs

**Pros:**
- Um fork publica sem tocar em DNS; quem quer domínio resolve em dois comandos.
- Trocar ou remover o domínio não gera commit.
- O erro mais provável (DNS não propagado) é detectado antes do deploy.
- Nenhum domínio fica embutido no molde, então nenhum fork herda um nome alheio.

**Cons:**
- O hostname deixa de ser legível só olhando o repositório — é preciso
  `make status` ou a aba Variables. Mitigado por `make url` e `make status`.
- ERB com lógica dentro do YAML é menos óbvio que uma string literal. Mitigado
  pelo comentário no arquivo e pelo teste.
- A conferência de DNS depende de `getent`, `dig` ou `python3`. Sem nenhum dos
  três o script avisa e segue em frente, em vez de travar.
- `www` responde mas não redireciona para o domínio canônico. Aceitável aqui
  (a sessão demo é local ao navegador); num app real exigiria redirect 301.

## Alternatives Considered

- **Editar `proxy.host` no arquivo** (o caminho documentado no padrão de
  deploy): descartado pelos motivos acima — exige commit por participante e
  arrisca vazar um domínio para todos os forks.
- **Guardar o domínio como secret:** descartado — não é sensível, e como secret
  ficaria ilegível no painel, atrapalhando a conferência.
- **Detectar automaticamente se o domínio é apex e acrescentar `www`:**
  descartado — distinguir apex de subdomínio exige a lista de sufixos públicos
  (`exemplo.com.br` tem três rótulos e é apex; `app.exemplo.dev` tem três e não
  é). A lista separada por vírgula resolve sem heurística frágil.
- **Emitir o certificado por desafio DNS-01**, que dispensaria a ordem
  DNS-antes-do-deploy: descartado — exigiria credenciais de API do provedor de
  DNS de cada participante, muito mais setup do que o problema justifica.
