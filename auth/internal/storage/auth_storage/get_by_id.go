package auth_storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/artem13815/hr/auth/internal/domain"
	"github.com/jackc/pgx/v5"
	pkgerrors "github.com/pkg/errors"
)

func (s *AuthStorage) GetUserByID(ctx context.Context, userID uint64) (*domain.User, error) {
	var u domain.User
	err := s.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s, %s, %s, %s, %s
		FROM %s
		WHERE %s = $1
	`, idColumn, emailColumn, passwordHashColumn, roleColumn, createdAtColumn, tableName, idColumn),
		userID,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, pkgerrors.Wrap(err, "failed to get user by id")
	}

	return &u, nil
}
