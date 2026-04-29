package bootstrap

import (
	"fmt"

	"github.com/artem13815/hr/analysis/config"
	"github.com/artem13815/hr/analysis/internal/storage/analysis_storage"
)

func InitPGStorage(cfg *config.Config) *analysis_storage.AnalysisStorage {
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	storage, err := analysis_storage.NewAnalysisStorage(connectionString)
	if err != nil {
		panic(err)
	}
	return storage
}
