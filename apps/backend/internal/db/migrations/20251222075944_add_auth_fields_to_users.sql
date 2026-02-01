-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD COLUMN password VARCHAR(255) NOT NULL;
ALTER TABLE users ADD COLUMN role VARCHAR(255) DEFAULT 'User';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP COLUMN password;
ALTER TABLE users DROP COLUMN role;
-- +goose StatementEnd