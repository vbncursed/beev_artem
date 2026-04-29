package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/artem13815/hr/auth/internal/domain"
)

// issueTokens mints an access/refresh pair and persists the refresh-session
// row. All token-format work is delegated to the TokenIssuer port (default
// implementation: infrastructure/jwt). Use case stays free of jwt + crypto
// imports.
func (s *AuthService) issueTokens(ctx context.Context, userID uint64, email, role, userAgent, ip string) (*domain.AuthInfo, error) {
	accessToken, err := s.tokenIssuer.IssueAccess(userID, email, role)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	refreshToken, err := s.tokenIssuer.IssueRefresh()
	if err != nil {
		return nil, fmt.Errorf("issue refresh token: %w", err)
	}

	refreshHash := s.tokenIssuer.HashRefresh(refreshToken)
	refreshExp := time.Now().Add(s.refreshTTL)

	if err := s.sessionStorage.CreateSession(ctx, userID, refreshHash, refreshExp, userAgent, ip); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return &domain.AuthInfo{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
