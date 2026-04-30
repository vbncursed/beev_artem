package bootstrap

import (
	"github.com/artem13815/hr/vacancy/internal/infrastructure/persistence"
	"github.com/artem13815/hr/vacancy/internal/usecase"
)

func InitVacancyService(storage *persistence.VacancyStorage) *usecase.VacancyService {
	return usecase.NewVacancyService(storage)
}
