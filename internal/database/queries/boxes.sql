-- name: UpsertBox :one
INSERT INTO boxes (id, title, path, active)
VALUES (?, ?, ?, ?)
ON CONFLICT (id) DO UPDATE
SET title = excluded.title, path = excluded.path, active = excluded.active
RETURNING *;

-- name: GetBoxByID :one
SELECT * FROM boxes WHERE id = ?;

-- name: ListBoxes :many
SELECT * FROM boxes ORDER BY title;

-- name: ListBoxesByPath :many
SELECT * FROM boxes WHERE path = ?;

-- name: ListBoxesByActive :many
SELECT * FROM boxes WHERE active = ? ORDER BY title;

-- name: DeleteBox :exec
DELETE FROM boxes WHERE id = ?;
