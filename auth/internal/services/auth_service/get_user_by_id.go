package auth_service

import (
	"context"

	"github.com/artem13815/hr/auth/internal/domain"
)

func (s *AuthService) GetUserByID(ctx context.Context, userID uint64) (*domain.User, error) {
	user, err := s.authStorage.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}
