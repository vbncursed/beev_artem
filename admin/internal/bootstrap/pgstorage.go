package bootstrap

import (
	"fmt"

	"github.com/artem13815/hr/admin/config"
	"github.com/artem13815/hr/admin/internal/infrastructure/persistence"
)

// InitPGStorage builds the connection string and hands it to the storage
// layer, which owns the boot timeout and the pool ping. cfg.Database fields
// are validated by config.LoadConfig.
func InitPGStorage(cfg *config.Config) (*persistence.AdminStorage, error) {
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	storage, err := persistence.NewAdminStorage(connectionString)
	if err != nil {
		return nil, fmt.Errorf("init pg storage: %w", err)
	}
	return storage, nil
}
