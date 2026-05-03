package bootstrap

import (
	"github.com/artem13815/hr/admin/internal/infrastructure/persistence"
	"github.com/artem13815/hr/admin/internal/usecase"
)

func InitAdminService(storage *persistence.AdminStorage, authClient usecase.AuthClient) *usecase.AdminService {
	return usecase.NewAdminService(storage, authClient)
}
