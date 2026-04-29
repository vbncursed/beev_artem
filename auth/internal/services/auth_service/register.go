package auth_service

import (
	"context"

	"github.com/artem13815/hr/auth/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func (s *AuthService) Register(ctx context.Context, in domain.RegisterInput) (*domain.AuthInfo, error) {
	if err := validateAuthInput(in.Email, in.Password); err != nil {
		return nil, err
	}

	existingUser, err := s.authStorage.GetUserByEmail(ctx, in.Email)
	if err == nil && existingUser != nil {
		return nil, ErrEmailAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	userID, err := s.authStorage.CreateUser(ctx, in.Email, string(passwordHash))
	if err != nil {
		return nil, err
	}

	return s.issueTokens(ctx, userID, in.Email, domain.RoleUser, in.UserAgent, in.IP)
}
