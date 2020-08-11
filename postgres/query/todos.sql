-- name: GetAllTodos :many
SELECT * FROM todos ORDER BY id;

-- name: CreateTodo :one
INSERT INTO todos (name) VALUES ($1) RETURNING *;

-- name: GetTodoById :one
SELECT * FROM todos WHERE id = $1 LIMIT 1;

-- name: DeleteTodoById :exec
DELETE FROM todos WHERE id = $1;

-- name: UpdateTodo :one
UPDATE todos SET name = $2, completed = $3 WHERE id = $1 RETURNING *;