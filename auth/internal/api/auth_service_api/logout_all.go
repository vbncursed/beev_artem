package auth_service_api

import (
	"context"
	"errors"

	pb_models "github.com/artem13815/hr/auth/internal/pb/models"
	"github.com/artem13815/hr/auth/internal/services/auth_service"
	"google.golang.org/grpc/codes"
)

func (a *AuthServiceAPI) LogoutAll(ctx context.Context, req *pb_models.LogoutAllRequest) (*pb_models.LogoutResponse, error) {
	claims, ok := ClaimsFromContext(ctx)
	if !ok {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required. Invalid or missing JWT token.")
	}

	err := a.authService.LogoutAll(ctx, claims.UserID, req.GetRefreshToken())
	if err != nil {
		switch {
		case errors.Is(err, auth_service.ErrInvalidArgument):
			return nil, newFieldError(codes.InvalidArgument, ErrCodeMissingField, "refresh_token", "Refresh token is required.")
		case errors.Is(err, auth_service.ErrInvalidRefreshToken):
			return nil, newError(codes.Unauthenticated, ErrCodeInvalidToken, "Invalid refresh token.")
		default:
			if isDatabaseError(err) {
				return nil, newError(codes.Unavailable, ErrCodeServiceUnavailable, "Service temporarily unavailable. Please try again later.")
			}
			return nil, newError(codes.Internal, ErrCodeInternal, "An internal error occurred. Please try again later.")
		}
	}

	return &pb_models.LogoutResponse{
		Success: true,
		Message: "All sessions revoked successfully.",
	}, nil
}
