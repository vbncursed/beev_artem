package middleware

import (
	"context"
	"errors"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/artem13815/hr/analysis/internal/pb/auth_api"
)

// authValidator is the narrow surface of auth_api.AuthServiceClient that the
// interceptor needs. Defined here so tests (and any future caller) can swap
// the dependency without dragging in the full generated client.
type authValidator interface {
	ValidateAccessToken(ctx context.Context, in *auth_api.ValidateAccessTokenRequest, opts ...grpc.CallOption) (*auth_api.ValidateAccessTokenResponse, error)
}

// UserContext is the authenticated caller identity attached to each request
// by the auth interceptor. Handlers MUST read it via Get; never parse JWT or
// x-user-id metadata themselves — that would defeat the interceptor's
// guarantee that identity is verified.
type UserContext struct {
	UserID  uint64
	Role    string
	IsAdmin bool
}

// userCtxKey is unexported so identity can only be set inside this package.
// Nothing outside can fake claims via context.WithValue.
type userCtxKey struct{}

// Get returns the authenticated UserContext placed by the auth interceptor.
// ok=false means the request reached this handler without authentication —
// which should never happen in production (the chain rejects unauthenticated
// requests up front), so callers usually treat ok=false as an internal error.
func Get(ctx context.Context) (*UserContext, bool) {
	uc, ok := ctx.Value(userCtxKey{}).(*UserContext)
	return uc, ok
}

func set(ctx context.Context, uc *UserContext) context.Context {
	return context.WithValue(ctx, userCtxKey{}, uc)
}

// authValidationTimeout caps a ValidateAccessToken RPC. Each protected handler
// pays this latency once; the auth service's own DB lookup runs inside it.
const authValidationTimeout = 3 * time.Second

// UnaryRecoveryInterceptor turns panics into codes.Internal so a single buggy
// handler can't take the whole process down. Always declare it first in the
// chain — must be the outermost wrapper to catch downstream panics.
func UnaryRecoveryInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("rpc panic",
				"method", info.FullMethod,
				"panic", r,
				"stack", string(debug.Stack()),
			)
			err = status.Error(codes.Internal, "internal error")
		}
	}()
	return handler(ctx, req)
}

// StreamRecoveryInterceptor mirrors UnaryRecoveryInterceptor for streaming
// RPCs (UploadResume).
func StreamRecoveryInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("rpc panic",
				"method", info.FullMethod,
				"panic", r,
				"stack", string(debug.Stack()),
			)
			err = status.Error(codes.Internal, "internal error")
		}
	}()
	return handler(srv, ss)
}

// UnaryLoggingInterceptor records one access log per RPC. Centralising it
// here means handlers only need to emit slog calls for domain forensics that
// can't be derived from method+code (e.g. user-mismatch on a delete).
func UnaryLoggingInterceptor(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	slog.Info("rpc",
		"method", info.FullMethod,
		"code", status.Code(err).String(),
		"duration", time.Since(start),
	)
	return resp, err
}

func StreamLoggingInterceptor(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	start := time.Now()
	err := handler(srv, ss)
	slog.Info("rpc",
		"method", info.FullMethod,
		"code", status.Code(err).String(),
		"duration", time.Since(start),
	)
	return err
}

// UnaryAuthInterceptor validates the JWT on every RPC by calling the auth
// service. The resulting identity is attached to the context — handlers MUST
// read it via Get, never from the raw gRPC metadata, otherwise an attacker
// on the docker network could impersonate any user just by setting
// x-user-id.
func UnaryAuthInterceptor(authClient authValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		uc, err := authenticate(ctx, authClient)
		if err != nil {
			return nil, err
		}
		return handler(set(ctx, uc), req)
	}
}

// StreamAuthInterceptor mirrors UnaryAuthInterceptor for streaming RPCs. We
// wrap the ServerStream so handlers see the authenticated context via
// stream.Context().
func StreamAuthInterceptor(authClient authValidator) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		uc, err := authenticate(ss.Context(), authClient)
		if err != nil {
			return err
		}
		return handler(srv, &authedServerStream{ServerStream: ss, ctx: set(ss.Context(), uc)})
	}
}

// authedServerStream overrides Context so wrapped handlers receive the
// authenticated context instead of the raw incoming one.
type authedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (a *authedServerStream) Context() context.Context { return a.ctx }

// authenticate extracts the bearer token, calls auth.ValidateAccessToken,
// and returns the identity. It is the single source of truth for who the
// caller is — handlers must not parse metadata themselves.
//
// Errors here are deliberately bare status.Error rather than a structured
// detail: an unauthenticated caller has not yet established a session, so we
// avoid leaking the reason-coding scheme used inside the service.
func authenticate(ctx context.Context, authClient authValidator) (*UserContext, error) {
	token, err := bearerTokenFromContext(ctx)
	if err != nil {
		return nil, status.Error(codes.Unauthenticated, "Authentication required.")
	}

	callCtx, cancel := context.WithTimeout(ctx, authValidationTimeout)
	defer cancel()

	res, err := authClient.ValidateAccessToken(callCtx, &auth_api.ValidateAccessTokenRequest{AccessToken: token})
	if err != nil {
		// Auth being down should not look like "your token is bad" — surface
		// it as Unavailable so clients retry instead of forcing a re-login.
		if status.Code(err) == codes.Unavailable {
			return nil, status.Error(codes.Unavailable, "Auth service unavailable.")
		}
		return nil, status.Error(codes.Unauthenticated, "Invalid access token.")
	}
	if !res.GetValid() || res.GetUserId() == 0 {
		return nil, status.Error(codes.Unauthenticated, "Invalid access token.")
	}

	role := strings.ToLower(strings.TrimSpace(res.GetRole()))
	if role == "" {
		role = "user"
	}
	return &UserContext{
		UserID:  res.GetUserId(),
		Role:    role,
		IsAdmin: role == "admin",
	}, nil
}

func bearerTokenFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", errors.New("metadata not found")
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return "", errors.New("authorization header missing")
	}
	header := values[0]
	if scheme, value, ok := strings.Cut(header, " "); ok && strings.EqualFold(scheme, "bearer") {
		header = value
	}
	if header == "" {
		return "", errors.New("empty bearer token")
	}
	return header, nil
}
