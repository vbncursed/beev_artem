package bootstrap

import (
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/artem13815/hr/resume/config"
	"github.com/artem13815/hr/resume/internal/api/resume_service_api"
	"github.com/artem13815/hr/resume/internal/pb/resume_api"
)

func AppRun(api *resume_service_api.ResumeServiceAPI, cfg *config.Config) {
	if err := runGRPCServer(api, cfg); err != nil {
		panic(err)
	}
}

func runGRPCServer(api *resume_service_api.ResumeServiceAPI, cfg *config.Config) error {
	lis, err := net.Listen("tcp", cfg.Server.GRPCAddr)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	resume_api.RegisterResumeServiceServer(s, api)

	slog.Info("gRPC server listening", "addr", cfg.Server.GRPCAddr)
	return s.Serve(lis)
}
