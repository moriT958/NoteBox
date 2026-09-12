-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS box_paths (
    id     INTEGER PRIMARY KEY AUTOINCREMENT,
    box_id INTEGER NOT NULL REFERENCES boxes(id) ON DELETE CASCADE,
    path   TEXT    NOT NULL,
    UNIQUE (box_id, path)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS box_paths;
-- +goose StatementEnd
