package auth_service_api

import (
	"context"
	"errors"
	"log/slog"

	pb_models "github.com/artem13815/hr/auth/internal/pb/models"
	"github.com/artem13815/hr/auth/internal/services/auth_service"
	"google.golang.org/grpc/codes"
)

func (a *AuthServiceAPI) Me(ctx context.Context, _ *pb_models.MeRequest) (*pb_models.MeResponse, error) {
	claims, err := a.getClaimsFromContext(ctx, a.jwtSecret)
	if err != nil {
		slog.Info("Me", "status", "error", "error", "unauthorized")
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required. Invalid or missing JWT token.")
	}

	user, err := a.authService.GetUserByID(ctx, claims.UserID)
	if err != nil {
		slog.Info("Me", "status", "error", "user_id", claims.UserID, "error", err.Error())
		switch {
		case errors.Is(err, auth_service.ErrUserNotFound):
			return nil, newError(codes.NotFound, ErrCodeUnauthorized, "User not found.")
		default:
			if isDatabaseError(err) {
				return nil, newError(codes.Unavailable, ErrCodeServiceUnavailable, "Service temporarily unavailable. Please try again later.")
			}
			return nil, newError(codes.Internal, ErrCodeInternal, "An internal error occurred. Please try again later.")
		}
	}

	slog.Info("Me", "status", "success", "user_id", user.ID)
	return &pb_models.MeResponse{
		UserId: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}, nil
}
