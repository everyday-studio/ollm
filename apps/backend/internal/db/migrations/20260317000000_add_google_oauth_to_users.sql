-- +goose Up
-- +goose StatementBegin

-- Allow password to be NULL for social login users
ALTER TABLE users ALTER COLUMN password DROP NOT NULL;

-- Add google_id column for linking Google accounts
ALTER TABLE users ADD COLUMN google_id VARCHAR(255) UNIQUE;

-- Index for fast lookup by google_id
CREATE INDEX IF NOT EXISTS idx_users_google_id ON users (google_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_google_id;
ALTER TABLE users DROP COLUMN IF EXISTS google_id;
ALTER TABLE users ALTER COLUMN password SET NOT NULL;
-- +goose StatementEnd
