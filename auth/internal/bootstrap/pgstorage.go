package bootstrap

import (
	"fmt"
	"log"

	"github.com/artem13815/hr/auth/config"
	"github.com/artem13815/hr/auth/internal/storage/auth_storage"
)

func InitPGStorage(cfg *config.Config) *auth_storage.AuthStorage {
	connectionString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName, cfg.Database.SSLMode)

	storage, err := auth_storage.NewAuthStorage(connectionString)
	if err != nil {
		log.Panicf("failed to initialize database: %v", err)
	}

	return storage
}
