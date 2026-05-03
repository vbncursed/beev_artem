package usecase

import (
	"context"

	"github.com/artem13815/hr/admin/internal/domain"
)

// GetOverview returns the dashboard-level counters. Authorization is enforced
// at the transport-middleware level (admin-only interceptor), so usecase
// trusts that any caller reaching here is an admin.
func (s *AdminService) GetOverview(ctx context.Context) (*domain.SystemStats, error) {
	return s.storage.GetSystemStats(ctx)
}
