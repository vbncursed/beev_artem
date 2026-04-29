package bootstrap

import (
	"github.com/artem13815/hr/auth/config"
	"github.com/artem13815/hr/auth/internal/infrastructure/jwt"
	transport_grpc "github.com/artem13815/hr/auth/internal/transport/grpc"
	"github.com/artem13815/hr/auth/internal/usecase"
)

func InitAuthServiceAPI(
	authService *usecase.AuthService,
	validator *jwt.Validator,
	loginLimiter, registerLimiter, refreshLimiter transport_grpc.RateLimiter,
) *transport_grpc.AuthServiceAPI {
	return transport_grpc.NewAuthServiceAPI(
		authService,
		validator,
		loginLimiter,
		registerLimiter,
		refreshLimiter,
	)
}

// InitJWTValidator builds a Validator pinned to HS256 with `exp` required.
// Both the AuthServiceAPI (ValidateAccessToken handler) and the gRPC auth
// interceptor share the same instance — there is no benefit to separate
// validators and a single one keeps the wire format authoritative.
func InitJWTValidator(cfg *config.Config) *jwt.Validator {
	return jwt.NewValidator(cfg.Auth.JWTSecret)
}
