package usecase

//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i AnalysisStorage,Scorer -o ./mocks -s _mock.go -g

import (
	"context"

	"github.com/artem13815/hr/analysis/internal/domain"
	pb_multiagent "github.com/artem13815/hr/analysis/internal/pb/multiagent_api"
)

// AnalysisStorage is the persistence-driven port. Implemented by
// internal/infrastructure/persistence. Note that StartAnalysis is gone — the
// usecase now orchestrates the steps (load, score, save) so storage stays
// "dumb" and tests can drive each step independently.
type AnalysisStorage interface {
	NewID() (string, error)
	LoadResumeContext(ctx context.Context, resumeID string, requestUserID uint64, isAdmin bool) (*domain.ResumeContext, error)
	LoadVacancySkills(ctx context.Context, vacancyID string) ([]domain.VacancySkill, error)
	SaveAnalysis(ctx context.Context, in domain.SaveAnalysisInput) error
	GetAnalysis(ctx context.Context, analysisID string, requestUserID uint64, isAdmin bool) (*domain.Analysis, error)
	ListCandidatesByVacancy(ctx context.Context, in domain.ListCandidatesByVacancyInput) (*domain.ListCandidatesByVacancyResult, error)
	UpdateAIDecision(ctx context.Context, analysisID string, ai domain.AIDecision) error
}

// Scorer is the heuristic / LLM scoring port. Pure compute, no I/O.
// Implemented today by internal/infrastructure/scorer (keyword heuristic);
// swappable for an LLM-backed adapter without touching usecase or storage.
type Scorer interface {
	Score(resumeText string, skills []domain.VacancySkill) domain.AnalysisPayload
}

type AnalysisService struct {
	storage          AnalysisStorage
	scorer           Scorer
	multiAgentClient pb_multiagent.MultiAgentServiceClient
}

func NewAnalysisService(storage AnalysisStorage, scorer Scorer, multiAgentClient pb_multiagent.MultiAgentServiceClient) *AnalysisService {
	return &AnalysisService{
		storage:          storage,
		scorer:           scorer,
		multiAgentClient: multiAgentClient,
	}
}
