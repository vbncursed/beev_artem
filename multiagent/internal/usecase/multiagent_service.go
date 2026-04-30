package usecase

//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i DecisionStorage -o ./mocks -s _mock.go -g

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/artem13815/hr/multiagent/internal/domain"
)

// DecisionStorage is the persistence port. Implementations live under
// internal/infrastructure/persistence — usecase MUST NOT import pgx, JSON
// codecs, or anything else transport-specific.
type DecisionStorage interface {
	StoreDecision(ctx context.Context, req domain.DecisionRequest, resp *domain.DecisionResponse) error
}

// ErrInvalidArgument is returned when the caller passes an empty request.
// Transport layer maps it to codes.InvalidArgument.
var ErrInvalidArgument = errors.New("invalid argument")

type MultiAgentService struct {
	storage DecisionStorage
}

func NewMultiAgentService(storage DecisionStorage) *MultiAgentService {
	return &MultiAgentService{storage: storage}
}

// GenerateDecision applies the heuristic decision rules and persists the
// (request, response) pair as an audit row. Replacing the heuristic with a
// real LLM later means swapping this method's body — the storage and
// transport contracts are stable.
func (s *MultiAgentService) GenerateDecision(ctx context.Context, req domain.DecisionRequest) (*domain.DecisionResponse, error) {
	if req.Model == "" && req.MatchScore == 0 && len(req.CandidateSkills) == 0 && len(req.MissingSkills) == 0 && req.ResumeText == "" {
		return nil, ErrInvalidArgument
	}

	matchScore := req.MatchScore
	missingCount := len(req.MissingSkills)

	recommendation := "no"
	confidence := float32(0.55)
	switch {
	case matchScore >= 75 && missingCount <= 1:
		recommendation = "hire"
		confidence = 0.86
	case matchScore >= 45:
		recommendation = "maybe"
		confidence = 0.68
	}

	rationale := fmt.Sprintf("score=%.2f, missing_skills=%d", matchScore, missingCount)
	feedback := "Усилить резюме по недостающим навыкам и добавить измеримые результаты проектов."
	if missingCount > 0 {
		feedback = "Добавьте подтверждение навыков: " + strings.Join(req.MissingSkills, ", ")
	}

	soft := "Коммуникация и командная работа требуют отдельного интервью и кейс-оценки."

	agents := []domain.AgentResult{
		{
			AgentName:      "ExtractorAgent",
			Summary:        "Профиль кандидата выделен из резюме и контекста скоринга.",
			StructuredJSON: `{"source":"resume_text+score_context"}`,
			Confidence:     0.82,
		},
		{
			AgentName:      "DecisionAgent",
			Summary:        "Рекомендация построена на score и разрывах must-have навыков.",
			StructuredJSON: fmt.Sprintf(`{"match_score":%.2f,"missing_count":%d}`, matchScore, missingCount),
			Confidence:     confidence,
		},
	}

	resp := &domain.DecisionResponse{
		HRRecommendation:  recommendation,
		Confidence:        confidence,
		HRRationale:       rationale,
		CandidateFeedback: feedback,
		SoftSkillsNotes:   soft,
		AgentResults:      agents,
		RawTrace:          "heuristic-multiagent-v1",
		CreatedAt:         time.Now(),
	}

	if err := s.storage.StoreDecision(ctx, req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
