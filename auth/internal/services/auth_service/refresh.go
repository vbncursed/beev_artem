package auth_service

import (
	"context"
	"time"

	"github.com/artem13815/hr/auth/internal/domain"
)

// Refresh rotates a refresh token. The old refresh token is consumed atomically
// (GETDEL) before any new tokens are issued, so two concurrent calls with the
// same refresh token cannot both succeed — the loser sees ErrInvalidRefreshToken.
func (s *AuthService) Refresh(ctx context.Context, in domain.RefreshInput) (*domain.AuthInfo, error) {
	if in.RefreshToken == "" {
		return nil, ErrInvalidArgument
	}

	refreshHash := tokenToHash(in.RefreshToken)

	sess, err := s.sessionStorage.ConsumeSessionByRefreshHash(ctx, refreshHash)
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// From here on the session is already gone from storage. Any error path
	// forces the caller to re-authenticate, which is the safe default.

	if time.Now().After(sess.ExpiresAt) {
		return nil, ErrSessionExpired
	}

	user, err := s.authStorage.GetUserByID(ctx, sess.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	return s.issueTokens(ctx, user.ID, user.Email, user.Role, in.UserAgent, in.IP)
}
