package bootstrap

import (
	"github.com/artem13815/hr/vacancy/internal/services/vacancy_service"
	"github.com/artem13815/hr/vacancy/internal/storage/vacancy_storage"
)

func InitVacancyService(storage *vacancy_storage.VacancyStorage) *vacancy_service.VacancyService {
	return vacancy_service.NewVacancyService(storage)
}
