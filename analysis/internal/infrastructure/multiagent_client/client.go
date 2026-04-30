package multiagent_client

import (
	"fmt"
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/artem13815/hr/analysis/config"
	pb_multiagent "github.com/artem13815/hr/analysis/internal/pb/multiagent_api"
)

// New dials the multiagent service so analysis can request LLM-backed HR
// decisions during StartAnalysis. The connection is long-lived and
// multiplexed by gRPC over HTTP/2 — do not create a new client per request.
// The returned cleanup MUST be passed to bootstrap.AppRun so the conn is
// closed during graceful shutdown.
//
// Plaintext is acceptable inside the docker compose network; real production
// should put TLS or service-mesh mTLS underneath.
//
// cfg.MultiAgent.GRPCAddr is validated by config.LoadConfig.
func New(cfg *config.Config) (pb_multiagent.MultiAgentServiceClient, func(), error) {
	conn, err := grpc.NewClient(cfg.MultiAgent.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, fmt.Errorf("dial multiagent service: %w", err)
	}
	cleanup := func() {
		if err := conn.Close(); err != nil {
			slog.Warn("close multiagent client conn", "err", err)
		}
	}
	return pb_multiagent.NewMultiAgentServiceClient(conn), cleanup, nil
}
