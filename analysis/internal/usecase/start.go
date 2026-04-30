package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/artem13815/hr/analysis/internal/domain"
	pb_multiagent "github.com/artem13815/hr/analysis/internal/pb/multiagent_api"
)

// multiagentTimeout caps a GenerateDecision RPC. Keep separate from
// resume/auth interceptor timeouts because LLM calls take longer than a
// token-validation lookup.
const multiagentTimeout = 5 * time.Second

// StartAnalysis is the orchestrator: it loads the resume + vacancy slice,
// runs the pure-compute scorer, persists the analysis row, and (if
// requested) fans out to the multiagent service to overwrite the AI
// decision. Each step has its own port so tests can drive them in
// isolation.
func (s *AnalysisService) StartAnalysis(ctx context.Context, in domain.StartAnalysisInput) (*domain.StartAnalysisResult, error) {
	if in.RequestUserID == 0 || strings.TrimSpace(in.ResumeID) == "" {
		return nil, ErrInvalidArgument
	}

	rc, err := s.storage.LoadResumeContext(ctx, in.ResumeID, in.RequestUserID, in.IsAdmin)
	if err != nil {
		return nil, err
	}
	if rc == nil {
		return nil, ErrNotFound
	}

	vacancyID := rc.VacancyID
	if v := strings.TrimSpace(in.VacancyID); v != "" {
		vacancyID = v
	}
	if vacancyID == "" {
		return nil, ErrInvalidArgument
	}

	skills, err := s.storage.LoadVacancySkills(ctx, vacancyID)
	if err != nil {
		return nil, err
	}

	payload := s.scorer.Score(rc.ResumeText, skills)

	analysisID, err := s.storage.NewID()
	if err != nil {
		return nil, err
	}

	if err := s.storage.SaveAnalysis(ctx, domain.SaveAnalysisInput{
		AnalysisID:     analysisID,
		VacancyID:      vacancyID,
		CandidateID:    rc.CandidateID,
		ResumeID:       rc.ResumeID,
		VacancyVersion: rc.VacancyVersion,
		Status:         domain.StatusDone,
		Score:          payload.Score,
		Profile:        payload.Profile,
		Breakdown:      payload.Breakdown,
		AI:             payload.AI,
	}); err != nil {
		return nil, err
	}

	res := &domain.StartAnalysisResult{AnalysisID: analysisID, Status: domain.StatusQueued}

	if in.UseLLM && s.multiAgentClient != nil {
		s.refreshAIDecisionAsync(ctx, analysisID, in.RequestUserID, in.IsAdmin, payload)
	}

	return res, nil
}

// refreshAIDecisionAsync runs the heuristic-derived payload through the
// multiagent service and overwrites the saved AI decision if the call
// succeeds. Errors are intentionally swallowed: the heuristic decision is
// the authoritative fallback, so failing here just means the row keeps the
// heuristic AI it already has.
//
// Inlined into StartAnalysis flow synchronously today (not actually async)
// because we want the freshly-saved row visible to the next read. If this
// becomes too slow, lift it to a background worker fed by an outbox table.
func (s *AnalysisService) refreshAIDecisionAsync(ctx context.Context, analysisID string, requestUserID uint64, isAdmin bool, payload domain.AnalysisPayload) {
	maCtx, cancel := context.WithTimeout(ctx, multiagentTimeout)
	defer cancel()

	maResp, err := s.multiAgentClient.GenerateDecision(maCtx, &pb_multiagent.GenerateDecisionRequest{
		Model:             "qwen-chat",
		Mode:              pb_multiagent.AgentMode_AGENT_MODE_BALANCED,
		CandidateSkills:   payload.Profile.Skills,
		MissingSkills:     payload.Breakdown.MissingSkills,
		CandidateSummary:  payload.Profile.Summary,
		ScoreExplanation:  payload.Breakdown.Explanation,
		MatchScore:        payload.Score,
		VacancyMustHave:   payload.Breakdown.MissingSkills,
		VacancyNiceToHave: payload.Breakdown.MatchedSkills,
		ResumeText:        payload.Profile.Summary,
	})
	if err != nil || maResp == nil {
		return
	}

	agentResults := make([]domain.AgentResult, 0, len(maResp.GetAgentResults()))
	for _, item := range maResp.GetAgentResults() {
		agentResults = append(agentResults, domain.AgentResult{
			AgentName:      item.GetAgentName(),
			Summary:        item.GetSummary(),
			StructuredJSON: item.GetStructuredJson(),
			Confidence:     item.GetConfidence(),
		})
	}
	_ = s.storage.UpdateAIDecision(ctx, analysisID, domain.AIDecision{
		HRRecommendation:  maResp.GetHrRecommendation(),
		Confidence:        maResp.GetConfidence(),
		HRRationale:       maResp.GetHrRationale(),
		CandidateFeedback: maResp.GetCandidateFeedback(),
		SoftSkillsNotes:   maResp.GetSoftSkillsNotes(),
		AgentResults:      agentResults,
		RawTrace:          maResp.GetRawTrace(),
	})
}
