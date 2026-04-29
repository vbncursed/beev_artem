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

func (a *AuthServiceAPI) Register(ctx context.Context, req *pb_models.RegisterRequest) (*pb_models.AuthResponse, error) {
	ua, ip := clientMeta(ctx)

	if !a.registerLimiter.Allow(ctx, ip) {
		slog.Info("Register", "status", "rate_limited", "ip", ip)
		return nil, newError(codes.ResourceExhausted, ErrCodeRateLimitExceeded, "Too many registration attempts. Please try again later.")
	}

	res, err := a.authService.Register(ctx, domain.RegisterInput{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		UserAgent: ua,
		IP:        ip,
	})
	if err != nil {
		slog.Info("Register", "status", "error", "email", req.GetEmail(), "error", err.Error())
		switch {
		case errors.Is(err, auth_service.ErrInvalidEmail):
			return nil, newFieldError(codes.InvalidArgument, ErrCodeInvalidEmail, "email", "Invalid email format.")
		case errors.Is(err, auth_service.ErrInvalidPassword):
			return nil, newFieldError(codes.InvalidArgument, ErrCodeInvalidPassword, "password", "Password must be at least 8 characters with uppercase, lowercase, and digit.")
		case errors.Is(err, auth_service.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid email or password format.")
		case errors.Is(err, auth_service.ErrEmailAlreadyExists):
			return nil, newError(codes.AlreadyExists, ErrCodeEmailAlreadyExists, "An account with this email already exists.")
		default:
			if isDatabaseError(err) {
				return nil, newError(codes.Unavailable, ErrCodeServiceUnavailable, "Service temporarily unavailable. Please try again later.")
			}
			return nil, newError(codes.Internal, ErrCodeInternal, "An internal error occurred. Please try again later.")
		}
	}

	slog.Info("Register", "status", "success", "user_id", res.UserID, "email", req.GetEmail())
	return &pb_models.AuthResponse{
		UserId:       res.UserID,
		AccessToken:  res.AccessToken,
		RefreshToken: res.RefreshToken,
	}, nil
}
