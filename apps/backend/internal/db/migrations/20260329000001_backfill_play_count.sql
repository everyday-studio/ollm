-- +goose Up
-- +goose StatementBegin
UPDATE games SET play_count = (SELECT COUNT(*) FROM matches WHERE matches.game_id = games.id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
UPDATE games SET play_count = 0;
-- +goose StatementEnd
