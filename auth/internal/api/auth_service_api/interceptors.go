package auth_service_api

import (
	"context"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// claimsCtxKey is unexported so only this package can put/get values; nobody
// outside can fake claims via context.WithValue.
type claimsCtxKey struct{}

// ClaimsFromContext returns the parsed JWT claims placed by UnaryAuthInterceptor.
// ok is false for public RPCs (Login/Register/Refresh/ValidateAccessToken) and
// for any caller that bypassed the interceptor (which shouldn't happen in
// production but we don't trust silence).
func ClaimsFromContext(ctx context.Context) (*tokenClaims, bool) {
	c, ok := ctx.Value(claimsCtxKey{}).(*tokenClaims)
	return c, ok
}

func ctxWithClaims(ctx context.Context, c *tokenClaims) context.Context {
	return context.WithValue(ctx, claimsCtxKey{}, c)
}

// publicMethods are RPCs that do not require an Authorization header — either
// because they are how clients obtain a token (Login, Register, Refresh) or
// because they validate one passed in the request body (ValidateAccessToken,
// called by the gateway).
var publicMethods = map[string]struct{}{
	"Login":               {},
	"Register":            {},
	"Refresh":             {},
	"ValidateAccessToken": {},
}

func isPublicMethod(fullMethod string) bool {
	// fullMethod looks like "/auth.service.v1.AuthService/Login".
	idx := strings.LastIndex(fullMethod, "/")
	if idx < 0 {
		return false
	}
	_, ok := publicMethods[fullMethod[idx+1:]]
	return ok
}

// UnaryRecoveryInterceptor converts panics into codes.Internal so a single
// buggy handler can't take the whole process down. Always declare it first in
// the chain — it must be the outermost wrapper to catch downstream panics.
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

// UnaryLoggingInterceptor centralises per-RPC access logging. Handlers should
// only emit slog calls for *domain-level* forensics that this generic record
// cannot derive from method+code (e.g. user-mismatch on Logout).
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

// UnaryAuthInterceptor parses the JWT for non-public methods and attaches the
// resulting claims to the context. Public RPCs pass through untouched. If a
// non-public RPC arrives without a valid token, the handler is never called —
// the client gets a clean codes.Unauthenticated.
//
// Centralising auth here removes ~10 lines of boilerplate from every protected
// handler and guarantees no handler can accidentally skip the check.
func UnaryAuthInterceptor(jwtSecret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if isPublicMethod(info.FullMethod) {
			return handler(ctx, req)
		}

		tokenString, err := extractTokenFromContext(ctx)
		if err != nil {
			return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required. Invalid or missing JWT token.")
		}
		claims, err := parseTokenClaims(tokenString, jwtSecret)
		if err != nil {
			return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required. Invalid or missing JWT token.")
		}

		return handler(ctxWithClaims(ctx, claims), req)
	}
}
