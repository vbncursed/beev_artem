package auth_service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/artem13815/hr/auth/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

func (s *AuthService) issueTokens(ctx context.Context, userID uint64, email, role, userAgent, ip string) (*domain.AuthInfo, error) {
	accessToken, err := s.generateAccessToken(userID, email, role)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, err := generateRefreshToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	refreshHash := hashRefreshToken(refreshToken)

	refreshExp := time.Now().Add(s.refreshTTL)

	err = s.sessionStorage.CreateSession(ctx, userID, refreshHash, refreshExp, userAgent, ip)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &domain.AuthInfo{
		UserID:       userID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) generateAccessToken(userID uint64, email, role string) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":     userID,
		"user_id": userID,
		"email":   email,
		"role":    role,
		"iat":     now.Unix(),
		"exp":     now.Add(s.accessTTL).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func generateRefreshToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func hashRefreshToken(refreshToken string) []byte {
	hash := sha256.Sum256([]byte(refreshToken))
	return hash[:]
}
