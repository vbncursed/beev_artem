package grpc

import (
	"github.com/artem13815/hr/auth/internal/transport/middleware"
	"context"
	"errors"
	"log/slog"

	"github.com/artem13815/hr/auth/internal/domain"
	pb_models "github.com/artem13815/hr/auth/internal/pb/models"
	"github.com/artem13815/hr/auth/internal/usecase"
	"google.golang.org/grpc/codes"
)

func (a *AuthServiceAPI) UpdateUserRole(ctx context.Context, req *pb_models.UpdateUserRoleRequest) (*pb_models.UpdateUserRoleResponse, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required. Invalid or missing JWT token.")
	}

	adminUserID := claims.UserID
	targetUserID := req.GetUserId()
	newRole := req.GetRole()

	if newRole != domain.RoleUser && newRole != domain.RoleAdmin {
		return nil, newFieldError(codes.InvalidArgument, ErrCodeInvalidInput, "role", "Role must be 'user' or 'admin'.")
	}

	if err := a.authService.UpdateUserRole(ctx, adminUserID, targetUserID, newRole); err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidRole):
			return nil, newFieldError(codes.InvalidArgument, ErrCodeInvalidInput, "role", "Invalid role. Must be 'user' or 'admin'.")
		case errors.Is(err, usecase.ErrPermissionDenied):
			return nil, newError(codes.PermissionDenied, ErrCodeForbidden, "Only administrators can change user roles.")
		case errors.Is(err, usecase.ErrCannotChangeOwnRole):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Cannot change your own role.")
		default:
			if isDatabaseError(err) {
				return nil, newError(codes.Unavailable, ErrCodeServiceUnavailable, "Service temporarily unavailable. Please try again later.")
			}
			return nil, newError(codes.Internal, ErrCodeInternal, "An internal error occurred. Please try again later.")
		}
	}

	// Audit trail: role changes are admin actions, worth keeping a structured
	// record beyond the generic RPC log line.
	slog.Info("role updated",
		"admin_user_id", adminUserID,
		"target_user_id", targetUserID,
		"new_role", newRole,
	)

	return &pb_models.UpdateUserRoleResponse{
		Success: true,
		Message: "User role updated successfully.",
	}, nil
}
