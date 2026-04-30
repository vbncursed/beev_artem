package bootstrap

import (
	"fmt"

	"github.com/artem13815/hr/analysis/config"
	"github.com/artem13815/hr/analysis/internal/infrastructure/persistence"
)

// InitPGStorage builds the connection string and hands it to the storage
// layer, which owns the boot timeout, the pool ping, and the goose migration
// pass. cfg.Database fields are validated by config.LoadConfig.
func InitPGStorage(cfg *config.Config) (*persistence.AnalysisStorage, error) {
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	storage, err := persistence.NewAnalysisStorage(connectionString)
	if err != nil {
		return nil, fmt.Errorf("init pg storage: %w", err)
	}
	return storage, nil
}
