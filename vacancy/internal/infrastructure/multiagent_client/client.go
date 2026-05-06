package multiagent_client

import (
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/artem13815/hr/vacancy/config"
	pb "github.com/artem13815/hr/vacancy/internal/pb/multiagent_api"
)

// New dials the multiagent service so vacancy can request LLM-based role
// classification during Create/Update. The connection is long-lived and
// multiplexed by gRPC over HTTP/2 — never create one per request.
//
// The returned cleanup MUST be passed to bootstrap.AppRun so the conn is
// closed during graceful shutdown, after the gRPC server stops accepting
// requests (LIFO order matters: in-flight Create handlers must finish
// before we tear down the multiagent conn).
//
// Plaintext is acceptable inside the docker compose network; real
// production should put TLS or service-mesh mTLS underneath. Mirrors the
// analysis -> multiagent client (analysis/internal/infrastructure/
// multiagent_client) by design — same network, same threat model.
//
// cfg.MultiAgent.GRPCAddr is validated by config.LoadConfig so reaching
// this constructor with an empty addr would already be a boot error.
func New(cfg *config.Config) (pb.MultiAgentServiceClient, func(), error) {
	conn, err := grpc.NewClient(cfg.MultiAgent.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("dial multiagent service: %w", err)
	}
	cleanup := func() {
		if err := conn.Close(); err != nil {
			slog.Warn("close multiagent client conn", "err", err)
		}
	}
	return pb.NewMultiAgentServiceClient(conn), cleanup, nil
}
