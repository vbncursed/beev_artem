package bootstrap

import (
	"fmt"

	"github.com/artem13815/hr/auth/config"
	"github.com/artem13815/hr/auth/internal/infrastructure/auth_storage"
)

// InitPGStorage builds the connection string from config and constructs the
// pgxpool-backed storage. NewAuthStorage applies its own context-bounded
// timeout for the initial connect/ping/init-tables sequence (M5).
//
// Errors are returned to main, which logs them via slog and exits cleanly —
// the previous log.Panicf would dump a stack trace into the container logs
// and exit with a confusing signal status.
func InitPGStorage(cfg *config.Config) (*auth_storage.AuthStorage, error) {
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	storage, err := auth_storage.NewAuthStorage(connectionString)
	if err != nil {
		return nil, fmt.Errorf("init pg storage: %w", err)
	}
	return storage, nil
}
