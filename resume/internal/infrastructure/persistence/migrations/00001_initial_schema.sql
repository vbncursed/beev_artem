-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS candidates (
    id            VARCHAR(64)  PRIMARY KEY,
    vacancy_id    VARCHAR(64)  NOT NULL,
    owner_user_id BIGINT       NOT NULL,
    full_name     VARCHAR(255) NOT NULL,
    email         VARCHAR(255) NOT NULL DEFAULT '',
    phone         VARCHAR(64)  NOT NULL DEFAULT '',
    source        VARCHAR(64)  NOT NULL DEFAULT '',
    comment       TEXT         NOT NULL DEFAULT '',
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_candidates_owner_user_id ON candidates(owner_user_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_candidates_vacancy_id ON candidates(vacancy_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS resumes (
    id              VARCHAR(64)  PRIMARY KEY,
    candidate_id    VARCHAR(64)  NOT NULL,
    file_name       VARCHAR(255) NOT NULL,
    file_type       VARCHAR(16)  NOT NULL,
    file_size_bytes BIGINT       NOT NULL,
    storage_path    TEXT         NOT NULL,
    extracted_text  TEXT         NOT NULL DEFAULT '',
    file_data       BYTEA        NOT NULL,
    created_at      TIMESTAMP    NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_candidate FOREIGN KEY (candidate_id) REFERENCES candidates(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_resumes_candidate_id ON resumes(candidate_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS resumes;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS candidates;
-- +goose StatementEnd
