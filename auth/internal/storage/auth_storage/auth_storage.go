package auth_storage

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type AuthStorage struct {
	db *pgxpool.Pool
}

func NewAuthStorage(connString string) (*AuthStorage, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse config")
	}

	db, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to database")
	}

	storage := &AuthStorage{
		db: db,
	}

	err = storage.initTables()
	if err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *AuthStorage) initTables() error {
	sql := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			%s BIGSERIAL PRIMARY KEY,
			%s VARCHAR(255) UNIQUE NOT NULL,
			%s VARCHAR(255) NOT NULL,
			%s VARCHAR(50) NOT NULL DEFAULT 'user',
			%s TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`, tableName, idColumn, emailColumn, passwordHashColumn, roleColumn, createdAtColumn)

	_, err := s.db.Exec(context.Background(), sql)
	if err != nil {
		return errors.Wrap(err, "failed to init tables")
	}

	return nil
}
