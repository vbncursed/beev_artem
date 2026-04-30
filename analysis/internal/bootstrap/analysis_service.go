package bootstrap

import (
	"github.com/artem13815/hr/analysis/internal/infrastructure/persistence"
	"github.com/artem13815/hr/analysis/internal/infrastructure/scorer"
	pb_multiagent "github.com/artem13815/hr/analysis/internal/pb/multiagent_api"
	"github.com/artem13815/hr/analysis/internal/usecase"
)

func InitAnalysisService(storage *persistence.AnalysisStorage, multiAgentClient pb_multiagent.MultiAgentServiceClient) *usecase.AnalysisService {
	return usecase.NewAnalysisService(storage, scorer.New(), multiAgentClient)
}
