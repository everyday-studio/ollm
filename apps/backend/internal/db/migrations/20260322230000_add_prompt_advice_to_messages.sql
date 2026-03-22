-- +goose NO TRANSACTION

-- +goose Up
-- +goose StatementBegin
ALTER TABLE messages ADD COLUMN IF NOT EXISTS prompt_advice TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE messages DROP COLUMN IF EXISTS prompt_advice;
-- +goose StatementEnd
