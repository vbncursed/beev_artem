package grpc

import (
	"context"

	"github.com/artem13815/hr/admin/internal/domain"
	"github.com/artem13815/hr/admin/internal/pb/admin_api"
)

type adminService interface {
	GetOverview(ctx context.Context) (*domain.SystemStats, error)
	ListUsers(ctx context.Context) ([]domain.AdminUserView, error)
	UpdateRole(ctx context.Context, in domain.UpdateRoleInput) error
}

type AdminServiceAPI struct {
	admin_api.UnimplementedAdminServiceServer
	svc adminService
}

func NewAdminServiceAPI(svc adminService) *AdminServiceAPI {
	return &AdminServiceAPI{svc: svc}
}

var _ admin_api.AdminServiceServer = (*AdminServiceAPI)(nil)
