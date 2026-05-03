// Package auth_client dials the auth service. The same long-lived gRPC
// connection serves both ValidateAccessToken (used by the auth interceptor)
// and UpdateUserRole (used by the admin usecase via RoleUpdater).
package auth_client

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/artem13815/hr/admin/config"
	"github.com/artem13815/hr/admin/internal/pb/auth_api"
)

// roleUpdateTimeout caps a single auth.UpdateUserRole RPC so a slow auth
// service can't hang an admin dashboard request indefinitely.
const roleUpdateTimeout = 5 * time.Second

// New dials the auth service so admin can validate JWTs on every protected RPC
// and proxy role changes. The connection is long-lived and multiplexed by gRPC
// over HTTP/2 — do not create a new client per request. The returned cleanup
// MUST be passed to bootstrap.AppRun so the conn is closed during graceful
// shutdown.
//
// Plaintext is acceptable inside the docker compose network; production should
// put TLS or service-mesh mTLS underneath.
//
// cfg.Auth.GRPCAddr is already validated by config.LoadConfig.
func New(cfg *config.Config) (auth_api.AuthServiceClient, func(), error) {
	conn, err := grpc.NewClient(cfg.Auth.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("dial auth service: %w", err)
	}
	cleanup := func() {
		if err := conn.Close(); err != nil {
			slog.Warn("close auth client conn", "err", err)
		}
	}
	return auth_api.NewAuthServiceClient(conn), cleanup, nil
}

// RoleUpdater adapts auth_api.AuthServiceClient to the usecase.AuthClient port.
// Lives in the infrastructure package because it carries the wire-level concern
// (timeout + error wrapping) the usecase deliberately doesn't know about.
type RoleUpdater struct {
	client auth_api.AuthServiceClient
}

// NewRoleUpdater wraps an existing auth client. Reuses the same gRPC conn that
// the auth interceptor uses for ValidateAccessToken — one dial, two callers.
func NewRoleUpdater(client auth_api.AuthServiceClient) *RoleUpdater {
	return &RoleUpdater{client: client}
}

// UpdateUserRole proxies the role change with a bounded timeout.
func (r *RoleUpdater) UpdateUserRole(ctx context.Context, userID uint64, newRole string) error {
	callCtx, cancel := context.WithTimeout(ctx, roleUpdateTimeout)
	defer cancel()

	if _, err := r.client.UpdateUserRole(callCtx, &auth_api.UpdateUserRoleRequest{
		UserId:  userID,
		NewRole: newRole,
	}); err != nil {
		return fmt.Errorf("auth.UpdateUserRole: %w", err)
	}
	return nil
}
