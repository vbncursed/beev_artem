package grpc

import (
	"github.com/artem13815/hr/auth/internal/transport/middleware"
	"context"
	"errors"

	pb_models "github.com/artem13815/hr/auth/internal/pb/models"
	"github.com/artem13815/hr/auth/internal/usecase"
	"google.golang.org/grpc/codes"
)

func (a *AuthServiceAPI) Logout(ctx context.Context, req *pb_models.LogoutRequest) (*pb_models.LogoutResponse, error) {
	claims, ok := middleware.ClaimsFromContext(ctx)
	if !ok {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required. Invalid or missing JWT token.")
	}

	err := a.authService.Logout(ctx, claims.UserID, req.GetRefreshToken())
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidArgument):
			return nil, newFieldError(codes.InvalidArgument, ErrCodeMissingField, "refresh_token", "Refresh token is required.")
		case errors.Is(err, usecase.ErrInvalidRefreshToken):
			// Uniform response for "not found" and "owned by other user" — see
			// service.Logout for the rationale (info-leak prevention).
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
		Message: "Session revoked successfully.",
	}, nil
}
