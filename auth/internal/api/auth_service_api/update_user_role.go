package auth_service_api

import (
	"context"
	"errors"
	"log/slog"

	"github.com/artem13815/hr/auth/internal/domain"
	pb_models "github.com/artem13815/hr/auth/internal/pb/models"
	"github.com/artem13815/hr/auth/internal/services/auth_service"
	"google.golang.org/grpc/codes"
)

func (a *AuthServiceAPI) UpdateUserRole(ctx context.Context, req *pb_models.UpdateUserRoleRequest) (*pb_models.UpdateUserRoleResponse, error) {
	adminUserID, err := a.getUserIDFromContext(ctx, a.jwtSecret)
	if err != nil {
		slog.Info("UpdateUserRole", "status", "error", "error", "unauthorized")
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required. Invalid or missing JWT token.")
	}

	targetUserID := req.GetUserId()
	newRole := req.GetRole()

	if newRole != domain.RoleUser && newRole != domain.RoleAdmin {
		slog.Info("UpdateUserRole", "status", "error", "admin_user_id", adminUserID, "target_user_id", targetUserID, "role", newRole, "error", "invalid_role")
		return nil, newFieldError(codes.InvalidArgument, ErrCodeInvalidInput, "role", "Role must be 'user' or 'admin'.")
	}

	err = a.authService.UpdateUserRole(ctx, adminUserID, targetUserID, newRole)
	if err != nil {
		slog.Info("UpdateUserRole", "status", "error", "admin_user_id", adminUserID, "target_user_id", targetUserID, "role", newRole, "error", err.Error())
		switch {
		case errors.Is(err, auth_service.ErrInvalidRole):
			return nil, newFieldError(codes.InvalidArgument, ErrCodeInvalidInput, "role", "Invalid role. Must be 'user' or 'admin'.")
		case errors.Is(err, auth_service.ErrPermissionDenied):
			return nil, newError(codes.PermissionDenied, ErrCodeForbidden, "Only administrators can change user roles.")
		case errors.Is(err, auth_service.ErrCannotChangeOwnRole):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Cannot change your own role.")
		default:
			if isDatabaseError(err) {
				return nil, newError(codes.Unavailable, ErrCodeServiceUnavailable, "Service temporarily unavailable. Please try again later.")
			}
			return nil, newError(codes.Internal, ErrCodeInternal, "An internal error occurred. Please try again later.")
		}
	}

	slog.Info("UpdateUserRole", "status", "success", "admin_user_id", adminUserID, "target_user_id", targetUserID, "new_role", newRole)
	return &pb_models.UpdateUserRoleResponse{
		Success: true,
		Message: "User role updated successfully.",
	}, nil
}
