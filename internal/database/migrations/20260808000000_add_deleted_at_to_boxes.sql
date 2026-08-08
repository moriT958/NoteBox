-- +goose Up
-- +goose StatementBegin
ALTER TABLE boxes ADD COLUMN deleted_at TIMESTAMP;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE boxes DROP COLUMN deleted_at;
-- +goose StatementEnd
