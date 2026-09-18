-- name: ListColumns :many
SELECT id, name, position FROM columns ORDER BY position;

-- name: ListTasks :many
SELECT id, column_id, title, description, position, created_at, updated_at
FROM tasks ORDER BY column_id, position;

-- name: GetTask :one
SELECT id, column_id, title, description, position, created_at, updated_at
FROM tasks WHERE id = $1;

-- name: CreateTask :one
INSERT INTO tasks (column_id, title, description)
VALUES ($1, $2, $3)
RETURNING id, column_id, title, description, position, created_at, updated_at;

-- name: UpdateTask :one
UPDATE tasks
SET column_id = $2,
    position  = $3,
    title     = $4,
    updated_at = now()
WHERE id = $1
RETURNING id, column_id, title, description, position, created_at, updated_at;

-- name: DeleteTask :exec
DELETE FROM tasks WHERE id = $1;
