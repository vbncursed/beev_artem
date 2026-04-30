package bootstrap

import (
	transport_grpc "github.com/artem13815/hr/analysis/internal/transport/grpc"
	"github.com/artem13815/hr/analysis/internal/usecase"
)

func InitAnalysisServiceAPI(service *usecase.AnalysisService) *transport_grpc.AnalysisServiceAPI {
	return transport_grpc.NewAnalysisServiceAPI(service)
}
