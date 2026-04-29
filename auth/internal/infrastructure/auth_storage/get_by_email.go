package auth_storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/artem13815/hr/auth/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (s *AuthStorage) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	var u domain.User
	err := s.db.QueryRow(ctx, fmt.Sprintf(`
		SELECT %s, %s, %s, %s, %s
		FROM %s
		WHERE %s = $1
	`, idColumn, emailColumn, passwordHashColumn, roleColumn, createdAtColumn, tableName, emailColumn),
		email,
	).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Role, &u.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return &u, nil
}
