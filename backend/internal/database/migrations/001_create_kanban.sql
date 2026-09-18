-- 001: Kanban board schema
-- A single-board demo: columns (Backlog, To Do, In Dev, Review, Done) and tasks.

CREATE TABLE IF NOT EXISTS columns (
    id          SERIAL PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    position    INT  NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS tasks (
    id          SERIAL PRIMARY KEY,
    column_id   INT  NOT NULL REFERENCES columns(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    position    INT  NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tasks_column ON tasks (column_id, position);

-- Seed the default columns for the demo board.
INSERT INTO columns (name, position) VALUES
    ('Backlog', 0),
    ('To Do',   1),
    ('In Dev',  2),
    ('Review',  3),
    ('Done',    4)
ON CONFLICT (name) DO NOTHING;
