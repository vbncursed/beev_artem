package bootstrap

import (
	"github.com/artem13815/hr/multiagent/internal/api/multiagent_service_api"
	"github.com/artem13815/hr/multiagent/internal/services/multiagent_service"
)

func InitMultiAgentServiceAPI(service *multiagent_service.MultiAgentService) *multiagent_service_api.MultiAgentServiceAPI {
	return multiagent_service_api.NewMultiAgentServiceAPI(service)
}
