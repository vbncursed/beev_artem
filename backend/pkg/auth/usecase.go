package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// AuthUseCase describes authentication/registration behavior.
type AuthUseCase interface {
	Register(ctx context.Context, email, password string) (AuthResult, error)
	Login(ctx context.Context, email, password string) (AuthResult, error)
}

type AuthResult struct {
	User  User
	Token string
}

type authService struct {
	repo   UserRepository
	tokens TokenGenerator
}

// NewAuthService returns default implementation of AuthUseCase.
func NewAuthService(repo UserRepository, tokens TokenGenerator) AuthUseCase {
	return &authService{repo: repo, tokens: tokens}
}

func (s *authService) Register(ctx context.Context, email, password string) (AuthResult, error) {
	if email == "" || password == "" {
		return AuthResult{}, ErrInvalidCredentials
	}

	// If user exists, fail fast (best-effort check)
	if _, err := s.repo.GetByEmail(ctx, email); err == nil {
		return AuthResult{}, ErrUserAlreadyExists
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, err
	}

	user := User{
		ID:           uuid.New(),
		Email:        email,
		PasswordHash: string(passwordHash),
		CreatedAt:    time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, user); err != nil {
		return AuthResult{}, err
	}
	token, err := s.tokens.Generate(ctx, user)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{User: user, Token: token}, nil
}

func (s *authService) Login(ctx context.Context, email, password string) (AuthResult, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return AuthResult{}, ErrInvalidCredentials
	}
	token, err := s.tokens.Generate(ctx, user)
	if err != nil {
		return AuthResult{}, err
	}
	return AuthResult{User: user, Token: token}, nil
}
