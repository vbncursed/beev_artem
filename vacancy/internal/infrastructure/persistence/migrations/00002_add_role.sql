-- +goose Up
-- +goose StatementBegin
ALTER TABLE vacancies
    ADD COLUMN IF NOT EXISTS role VARCHAR(64) NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE vacancies
    DROP COLUMN IF EXISTS role;
-- +goose StatementEnd
