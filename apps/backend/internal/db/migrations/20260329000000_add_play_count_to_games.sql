-- +goose Up
-- +goose StatementBegin
ALTER TABLE games ADD COLUMN play_count INTEGER NOT NULL DEFAULT 0;

CREATE OR REPLACE FUNCTION increment_game_play_count()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE games SET play_count = play_count + 1 WHERE id = NEW.game_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER increment_play_count_on_match_insert
    AFTER INSERT ON matches
    FOR EACH ROW
    EXECUTE FUNCTION increment_game_play_count();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS increment_play_count_on_match_insert ON matches;
DROP FUNCTION IF EXISTS increment_game_play_count();
ALTER TABLE games DROP COLUMN IF EXISTS play_count;
-- +goose StatementEnd
