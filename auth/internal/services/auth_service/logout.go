package auth_service

import (
	"context"
	"log/slog"
)

// Logout revokes a single refresh-token session. Returns ErrInvalidRefreshToken
// uniformly for "session not found" and "session belongs to another user" so
// that an attacker cannot probe whether a given refresh token exists by
// comparing error responses.
func (s *AuthService) Logout(ctx context.Context, userID uint64, refreshToken string) error {
	if refreshToken == "" {
		return ErrInvalidArgument
	}

	refreshHash := hashRefreshToken(refreshToken)

	sess, err := s.sessionStorage.GetSessionByRefreshHash(ctx, refreshHash)
	if err != nil {
		return ErrInvalidRefreshToken
	}

	if sess.UserID != userID {
		// Server-side audit: not exposed to the client (see uniform error
		// above), but we want to know if a token from one user is being
		// presented in another user's authenticated context.
		slog.Warn("logout: refresh token presented by different user",
			"caller_user_id", userID,
			"session_user_id", sess.UserID,
		)
		return ErrInvalidRefreshToken
	}

	return s.sessionStorage.RevokeSessionByRefreshHash(ctx, refreshHash)
}
