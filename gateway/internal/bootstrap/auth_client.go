package bootstrap

import (
	"github.com/artem13815/hr/gateway/config"
	"github.com/artem13815/hr/gateway/internal/infrastructure/auth_client"
	"github.com/artem13815/hr/gateway/internal/pb/auth_api"
)

// InitAuthClient builds the gateway's auth gRPC client. Mirrors the shape
// other services use so main can hand the cleanup to AppRun for LIFO
// teardown.
func InitAuthClient(cfg *config.Config) (auth_api.AuthServiceClient, func(), error) {
	return auth_client.New(cfg)
}
