-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN tag VARCHAR(5);

-- Populate existing rows to satisfy NOT NULL and UNIQUE constraints safely
UPDATE users SET tag = UPPER(SUBSTRING(MD5(id::text || random()::text) FROM 1 FOR 5)) WHERE tag IS NULL;

ALTER TABLE users ALTER COLUMN tag SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_tag_key UNIQUE (tag);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_tag_key;
ALTER TABLE users DROP COLUMN IF EXISTS tag;
-- +goose StatementEnd
