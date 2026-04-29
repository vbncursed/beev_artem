package bootstrap

import (
	"fmt"

	"github.com/artem13815/hr/vacancy/config"
	"github.com/artem13815/hr/vacancy/internal/storage/vacancy_storage"
)

func InitPGStorage(cfg *config.Config) *vacancy_storage.VacancyStorage {
	connectionString := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	storage, err := vacancy_storage.NewVacancyStorage(connectionString)
	if err != nil {
		panic(err)
	}
	return storage
}
