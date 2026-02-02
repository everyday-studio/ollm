-- +goose Up
-- +goose StatementBegin
-- Add NOT NULL constraint to existing timestamp columns for consistency
ALTER TABLE games 
  ALTER COLUMN created_at SET NOT NULL,
  ALTER COLUMN created_at SET DEFAULT CURRENT_TIMESTAMP,
  ALTER COLUMN updated_at SET NOT NULL,
  ALTER COLUMN updated_at SET DEFAULT CURRENT_TIMESTAMP;

-- Create trigger to automatically update updated_at timestamp on games table
-- Note: We reuse the update_updated_at_column() function created for users table
DROP TRIGGER IF EXISTS update_games_updated_at ON games;
CREATE TRIGGER update_games_updated_at
  BEFORE UPDATE ON games
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- Drop trigger
DROP TRIGGER IF EXISTS update_games_updated_at ON games;

-- Remove NOT NULL constraints (restore to original state)
ALTER TABLE games 
  ALTER COLUMN created_at DROP NOT NULL,
  ALTER COLUMN created_at SET DEFAULT NOW(),
  ALTER COLUMN updated_at DROP NOT NULL,
  ALTER COLUMN updated_at SET DEFAULT NOW();
-- +goose StatementEnd
