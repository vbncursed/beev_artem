package usecase

import (
	"context"

	"github.com/artem13815/hr/auth/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

func (s *AuthService) Login(ctx context.Context, in domain.LoginInput) (*domain.AuthInfo, error) {
	if err := validateAuthInput(in.Email, in.Password); err != nil {
		return nil, err
	}

	user, err := s.authStorage.GetUserByEmail(ctx, in.Email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(in.Password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokens(ctx, user.ID, user.Email, user.Role, in.UserAgent, in.IP)
}
