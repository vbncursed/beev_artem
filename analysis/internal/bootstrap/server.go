package bootstrap

import (
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/artem13815/hr/analysis/config"
	"github.com/artem13815/hr/analysis/internal/api/analysis_service_api"
	"github.com/artem13815/hr/analysis/internal/pb/analysis_api"
)

func AppRun(api *analysis_service_api.AnalysisServiceAPI, cfg *config.Config) {
	if err := runGRPCServer(api, cfg); err != nil {
		panic(err)
	}
}

func runGRPCServer(api *analysis_service_api.AnalysisServiceAPI, cfg *config.Config) error {
	lis, err := net.Listen("tcp", cfg.Server.GRPCAddr)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	analysis_api.RegisterAnalysisServiceServer(s, api)

	slog.Info("gRPC server listening", "addr", cfg.Server.GRPCAddr)
	return s.Serve(lis)
}
