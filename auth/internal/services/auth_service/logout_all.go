package auth_service

import (
	"context"
)

func (s *AuthService) LogoutAll(ctx context.Context, userID uint64, refreshToken string) error {
	if refreshToken == "" {
		return ErrInvalidArgument
	}

	refreshHash := tokenToHash(refreshToken)

	sess, err := s.sessionStorage.GetSessionByRefreshHash(ctx, refreshHash)
	if err != nil {
		return ErrInvalidRefreshToken
	}

	if sess.UserID != userID {
		return ErrPermissionDenied
	}

	return s.sessionStorage.RevokeAllSessionsByUserID(ctx, userID)
}
