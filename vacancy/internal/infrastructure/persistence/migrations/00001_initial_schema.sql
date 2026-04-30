-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS vacancies (
    id            VARCHAR(64)  PRIMARY KEY,
    owner_user_id BIGINT       NOT NULL,
    title         VARCHAR(255) NOT NULL,
    description   TEXT         NOT NULL,
    status        VARCHAR(32)  NOT NULL,
    version       INT          NOT NULL DEFAULT 1,
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_vacancies_owner_user_id ON vacancies(owner_user_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_vacancies_status ON vacancies(status);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS vacancy_skills (
    vacancy_id   VARCHAR(64)  NOT NULL,
    position     INT          NOT NULL,
    name         VARCHAR(255) NOT NULL,
    weight       REAL         NOT NULL,
    must_have    BOOLEAN      NOT NULL DEFAULT FALSE,
    nice_to_have BOOLEAN      NOT NULL DEFAULT FALSE,
    PRIMARY KEY (vacancy_id, position),
    CONSTRAINT fk_vacancy FOREIGN KEY (vacancy_id) REFERENCES vacancies(id) ON DELETE CASCADE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS vacancy_skills;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TABLE IF EXISTS vacancies;
-- +goose StatementEnd
