package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/artem13815/hr/analysis/internal/domain"
	pb_multiagent "github.com/artem13815/hr/analysis/internal/pb/multiagent_api"
)

func (s *AnalysisService) StartAnalysis(ctx context.Context, in domain.StartAnalysisInput) (*domain.StartAnalysisResult, error) {
	if in.RequestUserID == 0 || strings.TrimSpace(in.ResumeID) == "" {
		return nil, ErrInvalidArgument
	}

	res, err := s.storage.StartAnalysis(ctx, in)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrNotFound
	}

	if in.UseLLM && s.multiAgentClient != nil {
		analysis, err := s.storage.GetAnalysis(ctx, res.AnalysisID, in.RequestUserID, in.IsAdmin)
		if err == nil && analysis != nil {
			maCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			maResp, maErr := s.multiAgentClient.GenerateDecision(maCtx, &pb_multiagent.GenerateDecisionRequest{
				Model:              "qwen-chat",
				Mode:               pb_multiagent.AgentMode_AGENT_MODE_BALANCED,
				CandidateSkills:    analysis.Profile.Skills,
				MissingSkills:      analysis.Breakdown.MissingSkills,
				CandidateSummary:   analysis.Profile.Summary,
				ScoreExplanation:   analysis.Breakdown.Explanation,
				MatchScore:         analysis.MatchScore,
				VacancyMustHave:    analysis.Breakdown.MissingSkills,
				VacancyNiceToHave:  analysis.Breakdown.MatchedSkills,
				ResumeText:         analysis.Profile.Summary,
			})
			if maErr == nil && maResp != nil {
				agentResults := make([]domain.AgentResult, 0, len(maResp.GetAgentResults()))
				for _, item := range maResp.GetAgentResults() {
					agentResults = append(agentResults, domain.AgentResult{
						AgentName:      item.GetAgentName(),
						Summary:        item.GetSummary(),
						StructuredJSON: item.GetStructuredJson(),
						Confidence:     item.GetConfidence(),
					})
				}
				_ = s.storage.UpdateAIDecision(ctx, res.AnalysisID, domain.AIDecision{
					HRRecommendation:  maResp.GetHrRecommendation(),
					Confidence:        maResp.GetConfidence(),
					HRRationale:       maResp.GetHrRationale(),
					CandidateFeedback: maResp.GetCandidateFeedback(),
					SoftSkillsNotes:   maResp.GetSoftSkillsNotes(),
					AgentResults:      agentResults,
					RawTrace:          maResp.GetRawTrace(),
				})
			}
		}
	}

	return res, nil
}
