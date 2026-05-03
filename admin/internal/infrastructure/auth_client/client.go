// Package auth_client implements admin's AuthClient port via gRPC.
// admin holds an open gRPC.ClientConn to the auth service for the
// lifetime of the process; ValidateAccessToken (used by middleware)
// and UpdateUserRole (used by usecase) share the same conn.
package auth_client

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"

	"github.com/artem13815/hr/admin/internal/pb/auth_api"
)

// Client wraps the generated AuthServiceClient. The struct exists so
// tests can swap a fake while construction stays via NewClient.
type Client struct {
	conn   *grpc.ClientConn
	client auth_api.AuthServiceClient
}

// NewClient wires the gRPC connection. Uses NewClient (DialContext is
// deprecated in grpc-go 1.60+) so dialling is lazy — the first RPC
// call is when the network round-trip happens.
func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		conn:   conn,
		client: auth_api.NewAuthServiceClient(conn),
	}
}

// Raw returns the underlying generated client so middleware can call
// ValidateAccessToken without going through the AuthClient port.
func (c *Client) Raw() auth_api.AuthServiceClient {
	return c.client
}

// UpdateUserRole proxies the role change. 5-second deadline so a slow
// auth doesn't hang the dashboard request indefinitely.
func (c *Client) UpdateUserRole(ctx context.Context, userID uint64, newRole string) error {
	callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := c.client.UpdateUserRole(callCtx, &auth_api.UpdateUserRoleRequest{
		UserId:  userID,
		NewRole: newRole,
	})
	if err != nil {
		return fmt.Errorf("auth.UpdateUserRole: %w", err)
	}
	return nil
}
