package bootstrap

import (
	transport_grpc "github.com/artem13815/hr/vacancy/internal/transport/grpc"
	"github.com/artem13815/hr/vacancy/internal/usecase"
)

func InitVacancyServiceAPI(service *usecase.VacancyService) *transport_grpc.VacancyServiceAPI {
	return transport_grpc.NewVacancyServiceAPI(service)
}
