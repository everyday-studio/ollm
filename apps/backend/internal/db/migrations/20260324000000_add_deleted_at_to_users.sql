-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE NULL;
-- +goose StatementEnd

-- Create index concurrently without holding a lock
-- +goose NO TRANSACTION
-- +goose StatementBegin
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_users_deleted_at ON users (deleted_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_users_deleted_at;
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
-- +goose StatementEnd
