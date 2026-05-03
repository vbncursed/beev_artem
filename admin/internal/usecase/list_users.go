package usecase

import (
	"context"

	"github.com/artem13815/hr/admin/internal/domain"
)

// ListUsers returns every HR account with activity counters.
func (s *AdminService) ListUsers(ctx context.Context) ([]domain.AdminUserView, error) {
	return s.storage.ListUsers(ctx)
}
