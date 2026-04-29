package auth_service_api

import (
	"context"

	"github.com/artem13815/hr/auth/internal/domain"
	"github.com/artem13815/hr/auth/internal/pb/auth_api"
)

type authService interface {
	Register(ctx context.Context, in domain.RegisterInput) (*domain.AuthInfo, error)
	Login(ctx context.Context, in domain.LoginInput) (*domain.AuthInfo, error)
	Refresh(ctx context.Context, in domain.RefreshInput) (*domain.AuthInfo, error)
	GetUserByID(ctx context.Context, userID uint64) (*domain.User, error)
	Logout(ctx context.Context, userID uint64, refreshToken string) error
	LogoutAll(ctx context.Context, userID uint64, refreshToken string) error
	UpdateUserRole(ctx context.Context, adminUserID uint64, targetUserID uint64, newRole string) error
}

// AuthServiceAPI реализует grpc AuthServiceServer
type AuthServiceAPI struct {
	auth_api.UnimplementedAuthServiceServer
	authService     authService
	jwtSecret       string
	loginLimiter    RateLimiter
	registerLimiter RateLimiter
	refreshLimiter  RateLimiter
}

type RateLimiter interface {
	Allow(ctx context.Context, key string) bool
}

type denyAllLimiter struct{}

func (denyAllLimiter) Allow(ctx context.Context, key string) bool { return false }

func NewAuthServiceAPI(authService authService, jwtSecret string, loginLimiter, registerLimiter, refreshLimiter RateLimiter) *AuthServiceAPI {
	if loginLimiter == nil {
		loginLimiter = denyAllLimiter{}
	}
	if registerLimiter == nil {
		registerLimiter = denyAllLimiter{}
	}
	if refreshLimiter == nil {
		refreshLimiter = denyAllLimiter{}
	}
	return &AuthServiceAPI{
		authService:     authService,
		jwtSecret:       jwtSecret,
		loginLimiter:    loginLimiter,
		registerLimiter: registerLimiter,
		refreshLimiter:  refreshLimiter,
	}
}
