package analysis_storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalysisStorage struct {
	db *pgxpool.Pool
}

func NewAnalysisStorage(connectionString string) (*AnalysisStorage, error) {
	dbpool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	storage := &AnalysisStorage{db: dbpool}
	if err := storage.migrate(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to migrate tables: %w", err)
	}

	return storage, nil
}

func (s *AnalysisStorage) migrate(ctx context.Context) error {
	query := `
CREATE TABLE IF NOT EXISTS analyses (
  id VARCHAR(64) PRIMARY KEY,
  vacancy_id VARCHAR(64) NOT NULL,
  candidate_id VARCHAR(64) NOT NULL,
  resume_id VARCHAR(64) NOT NULL,
  vacancy_version INT NOT NULL DEFAULT 1,
  status VARCHAR(32) NOT NULL,
  match_score REAL NOT NULL DEFAULT 0,
  profile_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  breakdown_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  ai_json JSONB NOT NULL DEFAULT '{}'::jsonb,
  error_message TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_analyses_vacancy_id ON analyses(vacancy_id);
CREATE INDEX IF NOT EXISTS idx_analyses_candidate_id ON analyses(candidate_id);
CREATE INDEX IF NOT EXISTS idx_analyses_resume_id ON analyses(resume_id);
CREATE INDEX IF NOT EXISTS idx_analyses_created_at ON analyses(created_at DESC);
`

	_, err := s.db.Exec(ctx, query)
	return err
}
