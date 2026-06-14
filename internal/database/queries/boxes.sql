-- name: ListBoxes :many
SELECT * FROM boxes ORDER BY id;

-- name: CreateBox :one
INSERT INTO boxes (title, path)
VALUES (?, ?)
RETURNING *;

-- name: UpdateBox :one
UPDATE boxes
SET title = ?, path = ?
WHERE id = ?
RETURNING *;

-- name: DeleteBox :exec
DELETE FROM boxes
WHERE id = ?;
