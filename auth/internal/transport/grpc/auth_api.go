package grpc

import (
	"context"

	"github.com/artem13815/hr/auth/internal/domain"
	"github.com/artem13815/hr/auth/internal/infrastructure/jwt"
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

// RateLimiter is the consumer-side interface; the concrete implementation
// (rate_limit.Limiter) lives in infrastructure/rate_limit and is wired by
// bootstrap. The interface stays here so handlers can mock it in tests
// without dragging Redis in.
type RateLimiter interface {
	Allow(ctx context.Context, key string) bool
}

// AuthServiceAPI реализует grpc AuthServiceServer
type AuthServiceAPI struct {
	auth_api.UnimplementedAuthServiceServer
	authService     authService
	jwtValidator    *jwt.Validator
	loginLimiter    RateLimiter
	registerLimiter RateLimiter
	refreshLimiter  RateLimiter
}

// NewAuthServiceAPI requires non-nil limiters and a non-nil validator. The
// bootstrap layer (which always returns a real Redis-backed limiter via
// InitRateLimiters and a real jwt.Validator) guarantees this — passing nil
// is a programming error and we fail fast at construction time rather than
// nil-panicking on the first request.
func NewAuthServiceAPI(authService authService, jwtValidator *jwt.Validator, loginLimiter, registerLimiter, refreshLimiter RateLimiter) *AuthServiceAPI {
	if jwtValidator == nil {
		panic("auth_service_api: jwtValidator must be non-nil")
	}
	if loginLimiter == nil || registerLimiter == nil || refreshLimiter == nil {
		panic("auth_service_api: all rate limiters must be non-nil")
	}
	return &AuthServiceAPI{
		authService:     authService,
		jwtValidator:    jwtValidator,
		loginLimiter:    loginLimiter,
		registerLimiter: registerLimiter,
		refreshLimiter:  refreshLimiter,
	}
}
