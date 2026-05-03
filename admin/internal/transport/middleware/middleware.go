// Package middleware wires gRPC interceptors for admin: panic recovery,
// access logging, and **admin-only** auth (rejects any token whose claims
// don't carry role=admin). Owns UserContext + Get for handler-side
// identity reads.
package middleware

import (
	"context"
	"time"
)

// UserContext is the authenticated caller identity attached by the auth
// interceptor. Handlers MUST read it via Get; never parse JWT directly.
type UserContext struct {
	UserID  uint64
	Role    string
	IsAdmin bool
}

type userCtxKey struct{}

// Get returns the authenticated UserContext placed by the auth
// interceptor. ok=false should be impossible in production — the
// interceptor rejects unauthenticated requests up front.
func Get(ctx context.Context) (*UserContext, bool) {
	uc, ok := ctx.Value(userCtxKey{}).(*UserContext)
	return uc, ok
}

func set(ctx context.Context, uc *UserContext) context.Context {
	return context.WithValue(ctx, userCtxKey{}, uc)
}

const authValidationTimeout = 3 * time.Second
