-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS analyses (
    id              VARCHAR(64) PRIMARY KEY,
    vacancy_id      VARCHAR(64) NOT NULL,
    candidate_id    VARCHAR(64) NOT NULL,
    resume_id       VARCHAR(64) NOT NULL,
    vacancy_version INT         NOT NULL DEFAULT 1,
    status          VARCHAR(32) NOT NULL,
    match_score     REAL        NOT NULL DEFAULT 0,
    profile_json    JSONB       NOT NULL DEFAULT '{}'::jsonb,
    breakdown_json  JSONB       NOT NULL DEFAULT '{}'::jsonb,
    ai_json         JSONB       NOT NULL DEFAULT '{}'::jsonb,
    error_message   TEXT        NOT NULL DEFAULT '',
    created_at      TIMESTAMP   NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMP   NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_analyses_vacancy_id   ON analyses(vacancy_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_analyses_candidate_id ON analyses(candidate_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_analyses_resume_id    ON analyses(resume_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_analyses_created_at   ON analyses(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS analyses;
-- +goose StatementEnd
