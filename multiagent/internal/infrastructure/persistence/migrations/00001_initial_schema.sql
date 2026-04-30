-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS multiagent_decisions (
    id            VARCHAR(64)  PRIMARY KEY,
    model         VARCHAR(128) NOT NULL,
    mode          INT          NOT NULL,
    request_json  JSONB        NOT NULL,
    response_json JSONB        NOT NULL,
    created_at    TIMESTAMP    NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_multiagent_created_at ON multiagent_decisions(created_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS multiagent_decisions;
-- +goose StatementEnd
