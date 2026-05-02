// Package middleware wires gRPC interceptors that every service handler
// runs through: panic recovery, structured access logging, and JWT-backed
// authentication. The package also owns the `UserContext` type that
// authenticated handlers read via `Get`.
package middleware

import (
	"context"
	"time"
)

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
