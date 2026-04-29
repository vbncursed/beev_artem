package auth_client

import (
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/artem13815/hr/resume/config"
	"github.com/artem13815/hr/resume/internal/pb/auth_api"
)

// New dials the auth service so resume can validate JWTs on every protected
// RPC. The connection is long-lived and multiplexed by gRPC over HTTP/2 —
// do not create a new client per request. The returned cleanup MUST be passed
// to bootstrap.AppRun so the conn is closed during graceful shutdown.
//
// Plaintext is acceptable inside the docker compose network; real production
// should put TLS or service-mesh mTLS underneath.
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
