package bootstrap

import (
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/artem13815/hr/multiagent/config"
	"github.com/artem13815/hr/multiagent/internal/api/multiagent_service_api"
	pb "github.com/artem13815/hr/multiagent/internal/pb/multiagent_api"
)

func AppRun(api *multiagent_service_api.MultiAgentServiceAPI, cfg *config.Config) {
	if err := runGRPCServer(api, cfg); err != nil {
		panic(err)
	}
}

func runGRPCServer(api *multiagent_service_api.MultiAgentServiceAPI, cfg *config.Config) error {
	lis, err := net.Listen("tcp", cfg.Server.GRPCAddr)
	if err != nil {
		return err
	}
	s := grpc.NewServer()
	pb.RegisterMultiAgentServiceServer(s, api)
	slog.Info("gRPC server listening", "addr", cfg.Server.GRPCAddr)
	return s.Serve(lis)
}
