package bootstrap

import (
	"fmt"

	"github.com/artem13815/hr/resume/config"
	"github.com/artem13815/hr/resume/internal/storage/resume_storage"
)

func InitPGStorage(cfg *config.Config) *resume_storage.ResumeStorage {
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	storage, err := resume_storage.NewResumeStorage(connectionString)
	if err != nil {
		panic(err)
	}
	return storage
}
