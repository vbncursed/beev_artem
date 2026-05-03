package bootstrap

import (
	"fmt"

	"github.com/artem13815/hr/admin/config"
	"github.com/artem13815/hr/admin/internal/infrastructure/persistence"
)

// InitPGStorage opens the read-only pool against the shared `hr` DB.
func InitPGStorage(cfg *config.Config) (*persistence.StatsStorage, error) {
	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)
	storage, err := persistence.NewStatsStorage(connString)
	if err != nil {
		return nil, fmt.Errorf("init stats storage: %w", err)
	}
	return storage, nil
}
