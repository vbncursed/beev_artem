package vacancy_storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type VacancyStorage struct {
	db *pgxpool.Pool
}

func NewVacancyStorage(connectionString string) (*VacancyStorage, error) {
	dbpool, err := pgxpool.New(context.Background(), connectionString)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	storage := &VacancyStorage{db: dbpool}
	if err := storage.migrate(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to migrate tables: %w", err)
	}

	return storage, nil
}

func (s *VacancyStorage) migrate(ctx context.Context) error {
	query := `
CREATE TABLE IF NOT EXISTS vacancies (
  id VARCHAR(64) PRIMARY KEY,
  owner_user_id BIGINT NOT NULL,
  title VARCHAR(255) NOT NULL,
  description TEXT NOT NULL,
  status VARCHAR(32) NOT NULL,
  version INT NOT NULL DEFAULT 1,
  created_at TIMESTAMP NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_vacancies_owner_user_id ON vacancies(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_vacancies_status ON vacancies(status);

CREATE TABLE IF NOT EXISTS vacancy_skills (
  vacancy_id VARCHAR(64) NOT NULL,
  position INT NOT NULL,
  name VARCHAR(255) NOT NULL,
  weight REAL NOT NULL,
  must_have BOOLEAN NOT NULL DEFAULT FALSE,
  nice_to_have BOOLEAN NOT NULL DEFAULT FALSE,
  PRIMARY KEY (vacancy_id, position),
  CONSTRAINT fk_vacancy FOREIGN KEY (vacancy_id) REFERENCES vacancies(id) ON DELETE CASCADE
);
`

	_, err := s.db.Exec(ctx, query)
	return err
}
