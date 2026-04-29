package bootstrap

import (
	"google.golang.org/grpc"
	"log/slog"
	"net"

	"github.com/artem13815/hr/auth/config"
	"github.com/artem13815/hr/auth/internal/api/auth_service_api"
	"github.com/artem13815/hr/auth/internal/pb/auth_api"
)

func AppRun(api *auth_service_api.AuthServiceAPI, cfg *config.Config) {
	if err := runGRPCServer(api, cfg); err != nil {
		panic(err)
	}
}

func runGRPCServer(api *auth_service_api.AuthServiceAPI, cfg *config.Config) error {
	lis, err := net.Listen("tcp", cfg.Server.GRPCAddr)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	auth_api.RegisterAuthServiceServer(s, api)

	slog.Info("gRPC server listening", "addr", cfg.Server.GRPCAddr)
	return s.Serve(lis)
}
