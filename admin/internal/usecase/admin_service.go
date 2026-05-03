package usecase

import (
	"context"
	"errors"

	"github.com/artem13815/hr/admin/internal/domain"
)

// StatsStorage is the persistence-driven port. Implemented by
// infrastructure/persistence (read-only pgx queries against the shared
// `hr` database). Admin operates at a layer above per-service ownership
// boundaries — documented as a deliberate exception in admin/README.md.
type StatsStorage interface {
	GetSystemStats(ctx context.Context) (*domain.SystemStats, error)
	ListUsers(ctx context.Context) ([]domain.AdminUserView, error)
}

// AuthClient wraps the gRPC call to auth.UpdateUserRole. Defined here
// (not in infrastructure) because usecase needs to mock it; the
// concrete implementation lives in infrastructure/auth_client.
type AuthClient interface {
	UpdateUserRole(ctx context.Context, userID uint64, newRole string) error
}

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrUnauthorized    = errors.New("unauthorized")
)

type AdminService struct {
	storage    StatsStorage
	authClient AuthClient
}

func NewAdminService(storage StatsStorage, authClient AuthClient) *AdminService {
	return &AdminService{storage: storage, authClient: authClient}
}

// GetOverview returns the dashboard-level counters. Authorization is
// enforced at the transport-middleware level (admin-only interceptor),
// so usecase trusts the IsAdmin flag we'd pass here in a richer audit
// model.
func (s *AdminService) GetOverview(ctx context.Context) (*domain.SystemStats, error) {
	return s.storage.GetSystemStats(ctx)
}

// ListUsers returns every HR account with activity counters.
func (s *AdminService) ListUsers(ctx context.Context) ([]domain.AdminUserView, error) {
	return s.storage.ListUsers(ctx)
}

// UpdateRole proxies the role change to auth.UpdateUserRole. The auth
// service does its own admin-only check, so this is defense in depth.
func (s *AdminService) UpdateRole(ctx context.Context, in domain.UpdateRoleInput) error {
	if in.TargetUserID == 0 {
		return ErrInvalidArgument
	}
	if in.NewRole != domain.RoleAdmin && in.NewRole != domain.RoleUser {
		return ErrInvalidArgument
	}
	if !in.IsAdmin {
		return ErrUnauthorized
	}
	return s.authClient.UpdateUserRole(ctx, in.TargetUserID, in.NewRole)
}
