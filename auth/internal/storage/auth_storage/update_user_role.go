package auth_storage

import (
	"context"
	"errors"
	"fmt"
)

func (s *AuthStorage) UpdateUserRole(ctx context.Context, userID uint64, role string) error {
	result, err := s.db.Exec(ctx, fmt.Sprintf(`
		UPDATE %s
		SET %s = $1
		WHERE %s = $2
	`, tableName, roleColumn, idColumn),
		role, userID,
	)

	if err != nil {
		return fmt.Errorf("update user role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return errors.New("user not found")
	}

	return nil
}
