package bootstrap

import (
	"github.com/artem13815/hr/vacancy/config"
	"github.com/artem13815/hr/vacancy/internal/infrastructure/multiagent_client"
	pb "github.com/artem13815/hr/vacancy/internal/pb/multiagent_api"
)

// InitMultiAgentClient is a thin wrapper around multiagent_client.New so
// main.go composes through bootstrap symmetrically with the auth client.
// Returns the gRPC client + a cleanup hook that AppRun runs LIFO during
// graceful shutdown.
func InitMultiAgentClient(cfg *config.Config) (pb.MultiAgentServiceClient, func(), error) {
	return multiagent_client.New(cfg)
}
