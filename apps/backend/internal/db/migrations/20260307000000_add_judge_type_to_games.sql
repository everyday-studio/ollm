-- +goose Up
-- +goose StatementBegin
ALTER TABLE games
ADD COLUMN judge_type VARCHAR(50) NOT NULL DEFAULT 'target_word' 
    CHECK (judge_type IN ('target_word', 'format_break', 'llm_judge')),
ADD COLUMN judge_condition TEXT NOT NULL DEFAULT '';

ALTER TABLE games
DROP COLUMN IF EXISTS target_word;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE games
ADD COLUMN target_word VARCHAR(255) NOT NULL DEFAULT '';

ALTER TABLE games
DROP COLUMN judge_type,
DROP COLUMN judge_condition;
-- +goose StatementEnd
