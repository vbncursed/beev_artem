package bootstrap

import (
	transport_grpc "github.com/artem13815/hr/admin/internal/transport/grpc"
	"github.com/artem13815/hr/admin/internal/usecase"
)

func InitAdminServiceAPI(service *usecase.AdminService) *transport_grpc.AdminServiceAPI {
	return transport_grpc.NewAdminServiceAPI(service)
}
