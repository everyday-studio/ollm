-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE games ADD COLUMN first_message TEXT NOT NULL DEFAULT '';

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE games DROP COLUMN first_message;
