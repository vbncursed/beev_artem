package grpc

import (
	"context"
	"errors"
	"log/slog"

	"github.com/artem13815/hr/auth/internal/domain"
	pb_models "github.com/artem13815/hr/auth/internal/pb/models"
	"github.com/artem13815/hr/auth/internal/usecase"
	"google.golang.org/grpc/codes"
)

func (a *AuthServiceAPI) Refresh(ctx context.Context, req *pb_models.RefreshRequest) (*pb_models.AuthResponse, error) {
	ua, ip := clientMeta(ctx)

	if !a.refreshLimiter.Allow(ctx, "ip:"+ip) {
		slog.Info("refresh rate limited", "ip", ip)
		return nil, newError(codes.ResourceExhausted, ErrCodeRateLimitExceeded, "Too many refresh attempts. Please try again later.")
	}

	res, err := a.authService.Refresh(ctx, domain.RefreshInput{
		RefreshToken: req.GetRefreshToken(),
		UserAgent:    ua,
		IP:           ip,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidArgument):
			return nil, newFieldError(codes.InvalidArgument, ErrCodeMissingField, "refresh_token", "Refresh token is required.")
		case errors.Is(err, usecase.ErrInvalidRefreshToken):
			return nil, newError(codes.Unauthenticated, ErrCodeInvalidToken, "Invalid refresh token.")
		case errors.Is(err, usecase.ErrSessionExpired):
			return nil, newError(codes.Unauthenticated, ErrCodeSessionExpired, "Session has expired. Please log in again.")
		case errors.Is(err, usecase.ErrSessionRevoked):
			return nil, newError(codes.Unauthenticated, ErrCodeSessionRevoked, "Session has been revoked. Please log in again.")
		default:
			if isDatabaseError(err) {
				return nil, newError(codes.Unavailable, ErrCodeServiceUnavailable, "Service temporarily unavailable. Please try again later.")
			}
			return nil, newError(codes.Internal, ErrCodeInternal, "An internal error occurred. Please try again later.")
		}
	}

	return &pb_models.AuthResponse{
		UserId:       res.UserID,
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}, nil
}
