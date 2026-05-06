package usecase

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/artem13815/hr/analysis/internal/domain"
	pb_multiagent "github.com/artem13815/hr/analysis/internal/pb/multiagent_api"
)

// multiagentTimeout caps a GenerateDecision RPC. Keep separate from
// resume/auth interceptor timeouts because LLM calls take longer than a
// token-validation lookup. Sized below the multiagent-side Yandex
// request_timeout (60s) so analysis surfaces the deadline first with a
// clear cancellation, instead of racing the upstream HTTP timeout.
const multiagentTimeout = 45 * time.Second

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
		s.refreshAIDecisionAsync(ctx, analysisID, rc.VacancyRole, rc.ResumeText, payload)
	}

	return res, nil
}

// refreshAIDecisionAsync runs the heuristic-derived payload through the
// multiagent service and overwrites the saved AI decision if the call
// succeeds. Errors are intentionally swallowed: the heuristic decision is
// the authoritative fallback, so failing here just means the row keeps the
// heuristic AI it already has.
//
// role is forwarded so multiagent can pick the prompt for this vacancy
// (empty -> default prompt; multiagent owns the fallback).
//
// Inlined into StartAnalysis flow synchronously today (not actually async)
// because we want the freshly-saved row visible to the next read. If this
// becomes too slow, lift it to a background worker fed by an outbox table.
func (s *AnalysisService) refreshAIDecisionAsync(ctx context.Context, analysisID string, role string, resumeText string, payload domain.AnalysisPayload) {
	// gRPC `string` fields must be valid UTF-8 — protobuf marshaling rejects
	// otherwise. PDF extraction (pdftotext, ledongthuc fallback) occasionally
	// emits stray bytes from glyphs without proper encoding. Sanitize every
	// resume-derived string before sending so one bad PDF doesn't kill the
	// LLM call.
	clean := func(s string) string { return strings.ToValidUTF8(s, "") }
	cleanSlice := func(in []string) []string {
		out := make([]string, len(in))
		for i, v := range in {
			out[i] = clean(v)
		}
		return out
	}
	resumeText = clean(resumeText)
	candidateSkills := cleanSlice(payload.Profile.Skills)
	missingSkills := cleanSlice(payload.Breakdown.MissingSkills)
	matchedSkills := cleanSlice(payload.Breakdown.MatchedSkills)
	candidateSummary := clean(payload.Profile.Summary)
	scoreExplanation := clean(payload.Breakdown.Explanation)
	slog.InfoContext(ctx, "multiagent call start",
		"analysis_id", analysisID,
		"role", role,
		"resume_len", len(resumeText))
	maCtx, cancel := context.WithTimeout(ctx, multiagentTimeout)
	defer cancel()

	maResp, err := s.multiAgentClient.GenerateDecision(maCtx, &pb_multiagent.GenerateDecisionRequest{
		Model:             "qwen-chat",
		Mode:              pb_multiagent.AgentMode_AGENT_MODE_BALANCED,
		Role:              role,
		CandidateSkills:   candidateSkills,
		MissingSkills:     missingSkills,
		CandidateSummary:  candidateSummary,
		ScoreExplanation:  scoreExplanation,
		MatchScore:        payload.Score,
		VacancyMustHave:   missingSkills,
		VacancyNiceToHave: matchedSkills,
		// Full resume text — date ranges, durations, and other context
		// regex extractors miss. Required for the model to compute
		// years_experience accurately.
		ResumeText:        resumeText,
	})
	if err != nil || maResp == nil {
		slog.WarnContext(ctx, "multiagent call failed", "analysis_id", analysisID, "err", err, "resp_nil", maResp == nil)
		return
	}
	slog.InfoContext(ctx, "multiagent call ok", "analysis_id", analysisID, "yoe", maResp.GetYearsExperience())

	agentResults := make([]domain.AgentResult, 0, len(maResp.GetAgentResults()))
	for _, item := range maResp.GetAgentResults() {
		agentResults = append(agentResults, domain.AgentResult{
			AgentName:      item.GetAgentName(),
			Summary:        item.GetSummary(),
			StructuredJSON: item.GetStructuredJson(),
			Confidence:     item.GetConfidence(),
		})
	}
	// Override the heuristic regex-based YOE with the LLM's read when the
	// model returned a positive value. 0 means the LLM couldn't infer it —
	// keep the heuristic in that case (better than nothing).
	if yoe := maResp.GetYearsExperience(); yoe > 0 {
		_ = s.storage.UpdateProfileYearsExperience(ctx, analysisID, yoe)
	}
	// Same idea for the candidate summary: the heuristic preview is the
	// raw resume head truncated to 320 runes; the LLM writes a real
	// 1–2 sentence profile blurb. Empty -> keep heuristic.
	if summary := strings.TrimSpace(maResp.GetCandidateSummary()); summary != "" {
		_ = s.storage.UpdateProfileSummary(ctx, analysisID, summary)
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
