package analysis_service

import (
	"context"

	"github.com/artem13815/hr/analysis/internal/domain"
	pb_multiagent "github.com/artem13815/hr/analysis/internal/pb/multiagent_api"
)

type AnalysisStorage interface {
	StartAnalysis(ctx context.Context, in domain.StartAnalysisInput) (*domain.StartAnalysisResult, error)
	GetAnalysis(ctx context.Context, analysisID string, requestUserID uint64, isAdmin bool) (*domain.Analysis, error)
	ListCandidatesByVacancy(ctx context.Context, in domain.ListCandidatesByVacancyInput) (*domain.ListCandidatesByVacancyResult, error)
	UpdateAIDecision(ctx context.Context, analysisID string, ai domain.AIDecision) error
}

type AnalysisService struct {
	storage          AnalysisStorage
	multiAgentClient pb_multiagent.MultiAgentServiceClient
}

func NewAnalysisService(storage AnalysisStorage, multiAgentClient pb_multiagent.MultiAgentServiceClient) *AnalysisService {
	return &AnalysisService{storage: storage, multiAgentClient: multiAgentClient}
}
