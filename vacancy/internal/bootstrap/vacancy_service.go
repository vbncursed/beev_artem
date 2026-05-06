package bootstrap

import (
	"github.com/artem13815/hr/vacancy/internal/infrastructure/multiagent_client"
	"github.com/artem13815/hr/vacancy/internal/infrastructure/persistence"
	pb "github.com/artem13815/hr/vacancy/internal/pb/multiagent_api"
	"github.com/artem13815/hr/vacancy/internal/usecase"
)

func InitVacancyService(storage *persistence.VacancyStorage, maClient pb.MultiAgentServiceClient) *usecase.VacancyService {
	classifier := multiagent_client.NewClassifier(maClient)
	return usecase.NewVacancyService(storage, classifier)
}
