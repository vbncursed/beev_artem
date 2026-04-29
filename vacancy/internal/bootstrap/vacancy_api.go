package bootstrap

import (
	"github.com/artem13815/hr/vacancy/internal/api/vacancy_service_api"
	"github.com/artem13815/hr/vacancy/internal/services/vacancy_service"
)

func InitVacancyServiceAPI(service *vacancy_service.VacancyService) *vacancy_service_api.VacancyServiceAPI {
	return vacancy_service_api.NewVacancyServiceAPI(service)
}
