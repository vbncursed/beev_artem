package grpc

import (
	"context"
	"errors"

	"github.com/artem13815/hr/admin/internal/domain"
	pb_models "github.com/artem13815/hr/admin/internal/pb/models"
	"github.com/artem13815/hr/admin/internal/transport/middleware"
	"github.com/artem13815/hr/admin/internal/usecase"
	"google.golang.org/grpc/codes"
)

func (a *AdminServiceAPI) PromoteUser(ctx context.Context, req *pb_models.PromoteUserRequest) (*pb_models.UpdateRoleResponse, error) {
	return a.changeRole(ctx, req.GetUserId(), domain.RoleAdmin)
}

func (a *AdminServiceAPI) DemoteUser(ctx context.Context, req *pb_models.DemoteUserRequest) (*pb_models.UpdateRoleResponse, error) {
	return a.changeRole(ctx, req.GetUserId(), domain.RoleUser)
}

func (a *AdminServiceAPI) changeRole(ctx context.Context, targetID uint64, newRole string) (*pb_models.UpdateRoleResponse, error) {
	uc, ok := middleware.Get(ctx)
	if !ok {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	err := a.svc.UpdateRole(ctx, domain.UpdateRoleInput{
		CallerUserID: uc.UserID,
		IsAdmin:      uc.IsAdmin,
		TargetUserID: targetID,
		NewRole:      newRole,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid role change request.")
		case errors.Is(err, usecase.ErrUnauthorized):
			return nil, newError(codes.PermissionDenied, ErrCodeForbidden, "Admin privileges required.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Role change failed.")
		}
	}

	return &pb_models.UpdateRoleResponse{
		UserId:  targetID,
		NewRole: newRole,
	}, nil
}
