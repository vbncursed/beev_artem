package bootstrap

import (
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/artem13815/hr/vacancy/config"
	"github.com/artem13815/hr/vacancy/internal/api/vacancy_service_api"
	"github.com/artem13815/hr/vacancy/internal/pb/vacancy_api"
)

func AppRun(api *vacancy_service_api.VacancyServiceAPI, cfg *config.Config) {
	if err := runGRPCServer(api, cfg); err != nil {
		panic(err)
	}
}

func runGRPCServer(api *vacancy_service_api.VacancyServiceAPI, cfg *config.Config) error {
	lis, err := net.Listen("tcp", cfg.Server.GRPCAddr)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	vacancy_api.RegisterVacancyServiceServer(s, api)

	slog.Info("gRPC server listening", "addr", cfg.Server.GRPCAddr)
	return s.Serve(lis)
}
