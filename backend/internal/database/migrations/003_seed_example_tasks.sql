-- 003: cards de exemplo para o quadro nao abrir vazio.
-- So insere: posicoes continuam depois do que a coluna ja tem, e o NOT EXISTS
-- por (coluna, titulo) torna a migration idempotente.

INSERT INTO tasks (column_id, title, description, position)
SELECT
    c.id,
    v.title,
    v.description,
    COALESCE((SELECT MAX(t.position) FROM tasks t WHERE t.column_id = c.id), -1) + v.pos
FROM (
    VALUES
        -- Backlog (10)
        ('Backlog', 1,  'Busca full-text nas tasks',                 'Usar pgroonga para buscar por título e descrição, com destaque nos termos encontrados.'),
        ('Backlog', 2,  'Filtrar o quadro por responsável',          'Selecionar uma pessoa e ver só as tasks dela, sem recarregar a página.'),
        ('Backlog', 3,  'Anexos nas tasks',                          'Upload de imagem e PDF, com miniatura no card e limite de tamanho.'),
        ('Backlog', 4,  'Histórico de movimentações',                'Registrar quem moveu a task, de qual coluna para qual e quando.'),
        ('Backlog', 5,  'Etiquetas coloridas',                       'Marcar tasks como bug, feature ou débito técnico, com filtro por etiqueta.'),
        ('Backlog', 6,  'Limite de WIP por coluna',                  'Avisar visualmente quando uma coluna passar do limite configurado.'),
        ('Backlog', 7,  'Atalhos de teclado',                        'Criar task com N, navegar com as setas, fechar o formulário com Esc.'),
        ('Backlog', 8,  'Exportar o quadro em CSV',                  'Baixar todas as tasks com coluna, título, descrição e data de criação.'),
        ('Backlog', 9,  'Acessibilidade do arrastar e soltar',       'Permitir mover uma task só pelo teclado, anunciando a mudança por leitor de tela.'),
        ('Backlog', 10, 'Paginação no histórico',                    'Carregar em blocos: hoje a consulta traz tudo de uma vez e cresce sem limite.'),

        -- To Do (3)
        ('To Do',   1,  'Editar a descrição da task',                'PATCH /api/tasks ignora o campo description — a query UpdateTask não atualiza a coluna.'),
        ('To Do',   2,  'Confirmar antes de excluir',                'O clique no X remove a task na hora, sem chance de desfazer.'),
        ('To Do',   3,  'Estado vazio com orientação',               'Uma coluna sem tasks mostra apenas "Vazio"; falta dizer o que fazer ali.'),

        -- In Dev (4)
        ('In Dev',  1,  'Reordenar tasks dentro da coluna',          'Hoje o arrastar só troca de coluna; falta definir a posição dentro da lista.'),
        ('In Dev',  2,  'Atualizar o contador sem recarregar',       'O número no cabeçalho da coluna só muda depois de recarregar a página.'),
        ('In Dev',  3,  'Quadro utilizável no celular',              'Abaixo de 640px as colunas viram abas em vez de rolagem horizontal.'),
        ('In Dev',  4,  'Testes de ponta a ponta do fluxo principal','Playwright cobrindo entrar, criar uma task e movê-la entre colunas.'),

        -- Review (2)
        ('Review',  1,  'Contraste dos tokens no tema escuro',       'Revisar texto secundário e bordas: alguns pares ficam abaixo de 4.5:1.'),
        ('Review',  2,  'Reportar erros 500 ao Sentry',              'Conferir que a mensagem capturada não carrega dado sensível da requisição.'),

        -- Done (4)
        ('Done',    1,  'Login demo sem depender do backend',        'Sessão criada no cliente; não chama mais uma rota que só existe em desenvolvimento.'),
        ('Done',    2,  'Deploy automático a cada push',             'O GitHub Actions provisiona as VMs e publica a imagem com Kamal.'),
        ('Done',    3,  'Servir a SPA pelo próprio binário Go',      'Uma imagem só: o React construído vai embutido junto do servidor, na porta 80.'),
        ('Done',    4,  'Domínio próprio com HTTPS',                 'Certificado Let''s Encrypt emitido automaticamente, sem passo manual.')
) AS v(column_name, pos, title, description)
JOIN columns c ON c.name = v.column_name
WHERE NOT EXISTS (
    SELECT 1 FROM tasks t
    WHERE t.column_id = c.id AND t.title = v.title
);
