package bootstrap

import (
	"time"

	"github.com/artem13815/hr/auth/config"
	"github.com/artem13815/hr/auth/internal/infrastructure/jwt"
	"github.com/artem13815/hr/auth/internal/usecase"
	"github.com/artem13815/hr/auth/internal/infrastructure/auth_storage"
	"github.com/artem13815/hr/auth/internal/infrastructure/session_storage"
)

func InitAuthService(authStorage *auth_storage.AuthStorage, sessionStorage *session_storage.SessionStorage, cfg *config.Config) *usecase.AuthService {
	issuer := jwt.NewIssuer(
		cfg.Auth.JWTSecret,
		time.Duration(cfg.Auth.AccessTTLSeconds)*time.Second,
	)
	return usecase.NewAuthService(
		authStorage,
		sessionStorage,
		issuer,
		time.Duration(cfg.Auth.RefreshTTLSeconds)*time.Second,
		cfg.Auth.BcryptCost,
	)
}
