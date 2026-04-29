package auth_storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/artem13815/hr/auth/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (s *AuthStorage) CreateUser(ctx context.Context, email string, passwordHash string) (uint64, error) {
	var userID uint64
	err := s.db.QueryRow(ctx, fmt.Sprintf(`
		INSERT INTO %s (%s, %s, %s)
		VALUES ($1, $2, $3)
		RETURNING %s
	`, tableName, emailColumn, passwordHashColumn, roleColumn, idColumn),
		email, passwordHash, domain.RoleUser,
	).Scan(&userID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, fmt.Errorf("create user: %w", err)
		}
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			return 0, fmt.Errorf("email already exists: %w", err)
		}
		return 0, fmt.Errorf("create user: %w", err)
	}

	return userID, nil
}
