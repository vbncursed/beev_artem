package usecase

import (
	"context"

	"github.com/artem13815/hr/auth/internal/domain"
)

func (s *AuthService) UpdateUserRole(ctx context.Context, adminUserID uint64, targetUserID uint64, newRole string) error {
	if newRole != domain.RoleUser && newRole != domain.RoleAdmin {
		return ErrInvalidRole
	}

	adminUser, err := s.authStorage.GetUserByID(ctx, adminUserID)
	if err != nil {
		return err
	}
	if adminUser == nil {
		return ErrUserNotFound
	}
	if adminUser.Role != domain.RoleAdmin {
		return ErrPermissionDenied
	}

	if adminUserID == targetUserID {
		return ErrCannotChangeOwnRole
	}

	targetUser, err := s.authStorage.GetUserByID(ctx, targetUserID)
	if err != nil {
		return err
	}
	if targetUser == nil {
		return ErrUserNotFound
	}

	return s.authStorage.UpdateUserRole(ctx, targetUserID, newRole)
}
