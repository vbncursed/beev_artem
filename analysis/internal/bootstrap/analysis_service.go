package bootstrap

import (
	"github.com/artem13815/hr/analysis/config"
	"github.com/artem13815/hr/analysis/internal/pb/multiagent_api"
	"github.com/artem13815/hr/analysis/internal/services/analysis_service"
	"github.com/artem13815/hr/analysis/internal/storage/analysis_storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func InitAnalysisService(storage *analysis_storage.AnalysisStorage, cfg *config.Config) *analysis_service.AnalysisService {
	conn, err := grpc.NewClient(cfg.MultiAgent.GRPCAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	client := multiagent_api.NewMultiAgentServiceClient(conn)
	return analysis_service.NewAnalysisService(storage, client)
}
