package bootstrap

import (
	transport_grpc "github.com/artem13815/hr/resume/internal/transport/grpc"
	"github.com/artem13815/hr/resume/internal/usecase"
)

func InitResumeServiceAPI(service *usecase.ResumeService) *transport_grpc.ResumeServiceAPI {
	return transport_grpc.NewResumeServiceAPI(service)
}
