-- +goose Up
-- +goose StatementBegin
ALTER TABLE games 
    ADD COLUMN IF NOT EXISTS system_prompt TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS target_word VARCHAR(255) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE games 
    DROP COLUMN IF EXISTS system_prompt,
    DROP COLUMN IF EXISTS target_word;
-- +goose StatementEnd
