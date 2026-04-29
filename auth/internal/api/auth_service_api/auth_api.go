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

// NewAuthServiceAPI requires non-nil limiters. The bootstrap layer (which
// always returns a real Redis-backed limiter via InitRateLimiters) guarantees
// this — passing nil is a programming error and we fail fast at construction
// time rather than nil-panicking on the first request.
func NewAuthServiceAPI(authService authService, jwtSecret string, loginLimiter, registerLimiter, refreshLimiter RateLimiter) *AuthServiceAPI {
	if loginLimiter == nil || registerLimiter == nil || refreshLimiter == nil {
		panic("auth_service_api: all rate limiters must be non-nil")
	}
	return &AuthServiceAPI{
		authService:     authService,
		jwtSecret:       jwtSecret,
		loginLimiter:    loginLimiter,
		registerLimiter: registerLimiter,
		refreshLimiter:  refreshLimiter,
	}
}
