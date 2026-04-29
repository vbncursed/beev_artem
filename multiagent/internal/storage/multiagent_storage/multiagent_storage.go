package multiagent_storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type MultiAgentStorage struct {
	db *pgxpool.Pool
}

func NewMultiAgentStorage(connectionString string) (*MultiAgentStorage, error) {
	dbpool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	storage := &MultiAgentStorage{db: dbpool}
	if err := storage.migrate(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to migrate tables: %w", err)
	}
	return storage, nil
}

func (s *MultiAgentStorage) migrate(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
CREATE TABLE IF NOT EXISTS multiagent_decisions (
  id VARCHAR(64) PRIMARY KEY,
  model VARCHAR(128) NOT NULL,
  mode INT NOT NULL,
  request_json JSONB NOT NULL,
  response_json JSONB NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_multiagent_created_at ON multiagent_decisions(created_at DESC);
`)
	return err
}
