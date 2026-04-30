package usecase

//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i DecisionStorage -o ./mocks -s _mock.go -g

import (
	"context"
	"fmt"
	"strings"
	"time"

	pb "github.com/artem13815/hr/multiagent/internal/pb/multiagent_api"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DecisionStorage interface {
	StoreDecision(ctx context.Context, req *pb.GenerateDecisionRequest, resp *pb.GenerateDecisionResponse) error
}

type MultiAgentService struct {
	storage DecisionStorage
}

func NewMultiAgentService(storage DecisionStorage) *MultiAgentService {
	return &MultiAgentService{storage: storage}
}

func (s *MultiAgentService) GenerateDecision(ctx context.Context, req *pb.GenerateDecisionRequest) (*pb.GenerateDecisionResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("empty request")
	}

	matchScore := req.GetMatchScore()
	missingCount := len(req.GetMissingSkills())

	recommendation := "no"
	confidence := float32(0.55)
	if matchScore >= 75 && missingCount <= 1 {
		recommendation = "hire"
		confidence = 0.86
	} else if matchScore >= 45 {
		recommendation = "maybe"
		confidence = 0.68
	}

	rationale := fmt.Sprintf("score=%.2f, missing_skills=%d", matchScore, missingCount)
	feedback := "Усилить резюме по недостающим навыкам и добавить измеримые результаты проектов."
	if len(req.GetMissingSkills()) > 0 {
		feedback = "Добавьте подтверждение навыков: " + strings.Join(req.GetMissingSkills(), ", ")
	}

	soft := "Коммуникация и командная работа требуют отдельного интервью и кейс-оценки."

	agents := []*pb.AgentResult{
		{
			AgentName:      "ExtractorAgent",
			Summary:        "Профиль кандидата выделен из резюме и контекста скоринга.",
			StructuredJson: `{"source":"resume_text+score_context"}`,
			Confidence:     0.82,
		},
		{
			AgentName:      "DecisionAgent",
			Summary:        "Рекомендация построена на score и разрывах must-have навыков.",
			StructuredJson: fmt.Sprintf(`{"match_score":%.2f,"missing_count":%d}`, matchScore, missingCount),
			Confidence:     confidence,
		},
	}

	resp := &pb.GenerateDecisionResponse{
		HrRecommendation:  recommendation,
		Confidence:        confidence,
		HrRationale:       rationale,
		CandidateFeedback: feedback,
		SoftSkillsNotes:   soft,
		AgentResults:      agents,
		RawTrace:          "heuristic-multiagent-v1",
		CreatedAt:         timestamppb.New(time.Now()),
	}

	if err := s.storage.StoreDecision(ctx, req, resp); err != nil {
		return nil, err
	}

	return resp, nil
}
