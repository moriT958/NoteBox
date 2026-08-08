-- name: ListActiveBoxes :many
SELECT * FROM boxes WHERE deleted_at IS NULL ORDER BY id;

-- name: ListAllBoxes :many
SELECT * FROM boxes ORDER BY id;

-- name: CreateBox :one
INSERT INTO boxes (title, path)
VALUES (?, ?)
ON CONFLICT (path) DO UPDATE
SET title = excluded.title, deleted_at = NULL
WHERE boxes.deleted_at IS NOT NULL
RETURNING *;

-- name: UpdateBox :one
UPDATE boxes
SET title = ?, path = ?
WHERE id = ?
RETURNING *;

-- name: DeleteBox :exec
UPDATE boxes
SET deleted_at = CURRENT_TIMESTAMP
WHERE id = ?;
