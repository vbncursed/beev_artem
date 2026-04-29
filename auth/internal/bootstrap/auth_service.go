package bootstrap

import (
	"github.com/artem13815/hr/auth/config"
	"github.com/artem13815/hr/auth/internal/services/auth_service"
	"github.com/artem13815/hr/auth/internal/storage/auth_storage"
	"github.com/artem13815/hr/auth/internal/storage/session_storage"
)

func InitAuthService(authStorage *auth_storage.AuthStorage, sessionStorage *session_storage.SessionStorage, cfg *config.Config) *auth_service.AuthService {
	return auth_service.NewAuthService(
		authStorage,
		sessionStorage,
		cfg.Auth.JWTSecret,
		cfg.Auth.AccessTTLSeconds,
		cfg.Auth.RefreshTTLSeconds,
	)
}
