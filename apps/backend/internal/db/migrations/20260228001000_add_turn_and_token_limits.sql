-- +goose NO TRANSACTION
-- +goose Up
-- +goose StatementBegin
ALTER TABLE games 
    ADD COLUMN IF NOT EXISTS max_turns INTEGER NOT NULL DEFAULT 5;

ALTER TABLE matches
    ADD COLUMN IF NOT EXISTS max_turns INTEGER NOT NULL DEFAULT 5;

ALTER TABLE messages
    ADD COLUMN IF NOT EXISTS turn_count INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS token_count INTEGER NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE messages
    DROP COLUMN IF EXISTS turn_count,
    DROP COLUMN IF EXISTS token_count;

ALTER TABLE matches
    DROP COLUMN IF EXISTS max_turns;

ALTER TABLE games
    DROP COLUMN IF EXISTS max_turns;
-- +goose StatementEnd
