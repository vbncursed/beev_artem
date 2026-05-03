package usecase

import (
	"context"

	"github.com/artem13815/hr/admin/internal/domain"
)

// UpdateRole proxies the role change to auth.UpdateUserRole. The auth service
// performs its own admin-only check, so the IsAdmin guard here is defense in
// depth.
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
