package bootstrap

import (
	"fmt"

	"github.com/artem13815/hr/multiagent/config"
	"github.com/artem13815/hr/multiagent/internal/storage/multiagent_storage"
)

func InitPGStorage(cfg *config.Config) *multiagent_storage.MultiAgentStorage {
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	storage, err := multiagent_storage.NewMultiAgentStorage(connectionString)
	if err != nil {
		panic(err)
	}
	return storage
}
