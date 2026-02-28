-- +goose NO TRANSACTION
-- +goose Up
-- +goose StatementBegin
ALTER TABLE matches DROP CONSTRAINT IF EXISTS matches_status_check;
ALTER TABLE matches ADD CONSTRAINT matches_status_check CHECK (status IN ('active', 'won', 'lost', 'generating', 'resigned', 'expired', 'error'));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE matches DROP CONSTRAINT IF EXISTS matches_status_check;
ALTER TABLE matches ADD CONSTRAINT matches_status_check CHECK (status IN ('active', 'won', 'lost', 'resigned', 'expired', 'error'));
-- +goose StatementEnd
