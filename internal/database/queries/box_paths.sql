-- name: ListBoxPaths :many
SELECT * FROM box_paths WHERE box_id = ? ORDER BY id;

-- name: ListAllBoxPaths :many
SELECT * FROM box_paths ORDER BY box_id, id;

-- name: AddBoxPath :one
INSERT INTO box_paths (box_id, path)
VALUES (?, ?)
RETURNING *;

-- name: RemoveBoxPath :exec
DELETE FROM box_paths WHERE id = ?;

-- name: RemoveBoxPathByPath :exec
DELETE FROM box_paths WHERE box_id = ? AND path = ?;
