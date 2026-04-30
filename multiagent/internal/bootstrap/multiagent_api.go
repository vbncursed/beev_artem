package bootstrap

import (
	transport_grpc "github.com/artem13815/hr/multiagent/internal/transport/grpc"
	"github.com/artem13815/hr/multiagent/internal/usecase"
)

func InitMultiAgentServiceAPI(service *usecase.MultiAgentService) *transport_grpc.MultiAgentServiceAPI {
	return transport_grpc.NewMultiAgentServiceAPI(service)
}
