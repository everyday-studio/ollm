-- +goose Up
-- +goose StatementBegin
-- Drop foreign key constraint first
ALTER TABLE games DROP CONSTRAINT IF EXISTS games_author_id_fkey;

-- Change users table ID to VARCHAR(26) for ULID
ALTER TABLE users
  ALTER COLUMN id TYPE VARCHAR(26);

-- Change games table IDs to VARCHAR(26) for ULID
ALTER TABLE games
  ALTER COLUMN id TYPE VARCHAR(26),
  ALTER COLUMN author_id TYPE VARCHAR(26);

-- Recreate foreign key constraint
ALTER TABLE games
  ADD CONSTRAINT games_author_id_fkey
  FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE SET NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Note: Rolling back will cause data loss as ULIDs cannot be converted back to integers
-- This down migration is provided for completeness but should be used with caution

-- Drop foreign key constraint
ALTER TABLE games DROP CONSTRAINT IF EXISTS games_author_id_fkey;

-- Revert games table to SERIAL
ALTER TABLE games
  ALTER COLUMN id TYPE INTEGER USING 0,
  ALTER COLUMN author_id TYPE INTEGER USING 0;

-- Revert users table to SERIAL
ALTER TABLE users
  ALTER COLUMN id TYPE INTEGER USING 0;

-- Recreate foreign key constraint
ALTER TABLE games
  ADD CONSTRAINT games_author_id_fkey
  FOREIGN KEY (author_id) REFERENCES users(id) ON DELETE SET NULL;
-- +goose StatementEnd
