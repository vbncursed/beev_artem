package analysis_service_api

import (
	"context"

	"github.com/artem13815/hr/analysis/internal/domain"
	"github.com/artem13815/hr/analysis/internal/pb/analysis_api"
)

type analysisService interface {
	StartAnalysis(ctx context.Context, in domain.StartAnalysisInput) (*domain.StartAnalysisResult, error)
	GetAnalysis(ctx context.Context, in domain.GetAnalysisInput) (*domain.Analysis, error)
	ListCandidatesByVacancy(ctx context.Context, in domain.ListCandidatesByVacancyInput) (*domain.ListCandidatesByVacancyResult, error)
}

type AnalysisServiceAPI struct {
	analysis_api.UnimplementedAnalysisServiceServer
	analysisService analysisService
}

func NewAnalysisServiceAPI(service analysisService) *AnalysisServiceAPI {
	return &AnalysisServiceAPI{analysisService: service}
}

var _ analysis_api.AnalysisServiceServer = (*AnalysisServiceAPI)(nil)
