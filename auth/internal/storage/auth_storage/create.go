package auth_storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/artem13815/hr/auth/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	pkgerrors "github.com/pkg/errors"
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
			return 0, pkgerrors.Wrap(err, "failed to create user")
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, pkgerrors.Wrap(err, "email already exists")
		}
		return 0, pkgerrors.Wrap(err, "failed to create user")
	}

	return userID, nil
}
