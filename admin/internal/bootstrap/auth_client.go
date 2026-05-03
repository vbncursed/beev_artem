package bootstrap

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/artem13815/hr/admin/config"
	"github.com/artem13815/hr/admin/internal/infrastructure/auth_client"
)

// InitAuthClient dials the auth service over the docker-net using
// plaintext gRPC. TLS is delegated to the service mesh in prod.
func InitAuthClient(cfg *config.Config) (*auth_client.Client, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		cfg.Auth.GRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("dial auth: %w", err)
	}
	return auth_client.NewClient(conn), conn, nil
}
