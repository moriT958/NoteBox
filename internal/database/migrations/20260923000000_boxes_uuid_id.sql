-- +goose Up
-- +goose StatementBegin
DROP TABLE IF EXISTS boxes;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE boxes (
    id     TEXT PRIMARY KEY,
    title  TEXT NOT NULL,
    path   TEXT NOT NULL UNIQUE,
    active BOOLEAN NOT NULL DEFAULT TRUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS boxes;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS boxes (
    id    INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT    NOT NULL,
    path  TEXT    NOT NULL UNIQUE,
    deleted_at TIMESTAMP
);
-- +goose StatementEnd
