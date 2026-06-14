-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS boxes (
    id    INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT    NOT NULL,
    path  TEXT    NOT NULL UNIQUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS boxes;
-- +goose StatementEnd
