package usecase

//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i AdminStorage,AuthClient -o ./mocks -s _mock.go -g

import (
	"context"

	"github.com/artem13815/hr/admin/internal/domain"
)

// AdminStorage is the persistence-driven port. Implemented by
// infrastructure/persistence (read-only pgx queries against the shared `hr`
// database). Admin operates at a layer above per-service ownership boundaries
// — documented as a deliberate exception in admin/README.md.
type AdminStorage interface {
	GetSystemStats(ctx context.Context) (*domain.SystemStats, error)
	ListUsers(ctx context.Context) ([]domain.AdminUserView, error)
}

// AuthClient wraps the gRPC call to auth.UpdateUserRole. Defined here (not in
// infrastructure) because usecase needs to mock it; the concrete adapter lives
// in infrastructure/auth_client.RoleUpdater.
type AuthClient interface {
	UpdateUserRole(ctx context.Context, userID uint64, newRole string) error
}

type AdminService struct {
	storage    AdminStorage
	authClient AuthClient
}

func NewAdminService(storage AdminStorage, authClient AuthClient) *AdminService {
	return &AdminService{storage: storage, authClient: authClient}
}
