package auth_service_api

import (
	"context"
	"errors"

	pb_models "github.com/artem13815/hr/auth/internal/pb/models"
	"github.com/artem13815/hr/auth/internal/services/auth_service"
	"google.golang.org/grpc/codes"
)

func (a *AuthServiceAPI) ValidateAccessToken(ctx context.Context, req *pb_models.ValidateAccessTokenRequest) (*pb_models.ValidateAccessTokenResponse, error) {
	token := req.GetAccessToken()
	if token == "" {
		return nil, newFieldError(codes.InvalidArgument, ErrCodeMissingField, "access_token", "Access token is required.")
	}

	claims, err := parseTokenClaims(token, a.jwtSecret)
	if err != nil {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Invalid access token.")
	}

	user, err := a.authService.GetUserByID(ctx, claims.UserID)
	if err != nil {
		switch {
		case errors.Is(err, auth_service.ErrUserNotFound):
			return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Invalid access token.")
		default:
			if isDatabaseError(err) {
				return nil, newError(codes.Unavailable, ErrCodeServiceUnavailable, "Service temporarily unavailable. Please try again later.")
			}
			return nil, newError(codes.Internal, ErrCodeInternal, "An internal error occurred. Please try again later.")
		}
	}

	return &pb_models.ValidateAccessTokenResponse{
		Valid:  true,
		UserId: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}, nil
}
