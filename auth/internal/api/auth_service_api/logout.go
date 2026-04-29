package auth_service_api

import (
	"context"
	"errors"
	"log/slog"

	pb_models "github.com/artem13815/hr/auth/internal/pb/models"
	"github.com/artem13815/hr/auth/internal/services/auth_service"
	"google.golang.org/grpc/codes"
)

func (a *AuthServiceAPI) Logout(ctx context.Context, req *pb_models.LogoutRequest) (*pb_models.LogoutResponse, error) {
	userID, err := a.getUserIDFromContext(ctx, a.jwtSecret)
	if err != nil {
		slog.Info("Logout", "status", "error", "error", "unauthorized")
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required. Invalid or missing JWT token.")
	}

	err = a.authService.Logout(ctx, userID, req.GetRefreshToken())
	if err != nil {
		slog.Info("Logout", "status", "error", "user_id", userID, "error", err.Error())
		switch {
		case errors.Is(err, auth_service.ErrInvalidArgument):
			return nil, newFieldError(codes.InvalidArgument, ErrCodeMissingField, "refresh_token", "Refresh token is required.")
		case errors.Is(err, auth_service.ErrInvalidRefreshToken):
			return nil, newError(codes.Unauthenticated, ErrCodeInvalidToken, "Invalid refresh token.")
		case errors.Is(err, auth_service.ErrPermissionDenied):
			return nil, newError(codes.PermissionDenied, ErrCodeForbidden, "You can only revoke your own sessions.")
		default:
			if isDatabaseError(err) {
				return nil, newError(codes.Unavailable, ErrCodeServiceUnavailable, "Service temporarily unavailable. Please try again later.")
			}
			return nil, newError(codes.Internal, ErrCodeInternal, "An internal error occurred. Please try again later.")
		}
	}

	slog.Info("Logout", "status", "success", "user_id", userID)
	return &pb_models.LogoutResponse{
		Success: true,
		Message: "Session revoked successfully.",
	}, nil
}
