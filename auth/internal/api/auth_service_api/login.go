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

func (a *AuthServiceAPI) Login(ctx context.Context, req *pb_models.LoginRequest) (*pb_models.AuthResponse, error) {
	ua, ip := clientMeta(ctx)

	// Two buckets: per-IP (best effort — gateway must forward x-client-ip) and
	// per-email. The email bucket is what stops credential stuffing on a single
	// account regardless of how many IPs the attacker rotates through.
	if !a.loginLimiter.Allow(ctx, "ip:"+ip) || !a.loginLimiter.Allow(ctx, "email:"+emailRateKey(req.GetEmail())) {
		// rate-limit hits show up as ResourceExhausted in the generic RPC log;
		// add the email_hash so we can correlate repeated attempts on the same
		// account across IPs.
		slog.Info("login rate limited", "ip", ip, "email_hash", emailRateKey(req.GetEmail()))
		return nil, newError(codes.ResourceExhausted, ErrCodeRateLimitExceeded, "Too many login attempts. Please try again later.")
	}

	res, err := a.authService.Login(ctx, domain.LoginInput{
		Email:     req.GetEmail(),
		Password:  req.GetPassword(),
		UserAgent: ua,
		IP:        ip,
	})
	if err != nil {
		// Domain-level forensics: pair the email_hash with the failure mode so
		// we can spot stuffing/probing patterns. Generic method+code logging
		// happens in UnaryLoggingInterceptor.
		slog.Info("login failed", "email_hash", emailRateKey(req.GetEmail()), "error", err.Error())
		switch {
		case errors.Is(err, auth_service.ErrInvalidEmail):
			return nil, newFieldError(codes.InvalidArgument, ErrCodeInvalidEmail, "email", "Invalid email format.")
		case errors.Is(err, auth_service.ErrInvalidPassword):
			return nil, newFieldError(codes.InvalidArgument, ErrCodeInvalidPassword, "password", "Password must be at least 8 characters with uppercase, lowercase, and digit.")
		case errors.Is(err, auth_service.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid email or password format.")
		case errors.Is(err, auth_service.ErrInvalidCredentials):
			return nil, newError(codes.Unauthenticated, ErrCodeInvalidCredentials, "Invalid email or password.")
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
