package auth_service

import (
	"context"
	"log/slog"
)

// LogoutAll revokes every active session for the calling user. Same uniform-error
// rationale as Logout: do not leak whether the supplied refresh token exists.
func (s *AuthService) LogoutAll(ctx context.Context, userID uint64, refreshToken string) error {
	if refreshToken == "" {
		return ErrInvalidArgument
	}

	refreshHash := hashRefreshToken(refreshToken)

	sess, err := s.sessionStorage.GetSessionByRefreshHash(ctx, refreshHash)
	if err != nil {
		return ErrInvalidRefreshToken
	}

	if sess.UserID != userID {
		slog.Warn("logout_all: refresh token presented by different user",
			"caller_user_id", userID,
			"session_user_id", sess.UserID,
		)
		return ErrInvalidRefreshToken
	}

	return s.sessionStorage.RevokeAllSessionsByUserID(ctx, userID)
}
