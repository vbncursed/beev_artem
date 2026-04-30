package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/analysis/internal/domain"
	pb_multiagent "github.com/artem13815/hr/analysis/internal/pb/multiagent_api"
)

type StartAnalysisSuite struct{ baseSuite }

// happyResume is the joined slice the storage returns; tests start from this
// and override only the field they want to exercise.
func happyResume() *domain.ResumeContext {
	return &domain.ResumeContext{
		ResumeID:       "r-1",
		CandidateID:    "c-1",
		VacancyID:      "v-ctx",
		OwnerUserID:    7,
		ResumeText:     "go pgx grpc",
		VacancyVersion: 3,
	}
}

// happyPayload is the scorer's output stub. Score=80 so the orchestrator
// hits the "high score" branch in the StartAnalysisResult and so any
// downstream multiagent fan-out can use it as input.
func happyPayload() domain.AnalysisPayload {
	return domain.AnalysisPayload{
		Score:     80,
		Profile:   domain.CandidateProfile{Skills: []string{"go"}},
		Breakdown: domain.ScoreBreakdown{MatchedSkills: []string{"go"}, MissingSkills: []string{"k8s"}},
		AI:        domain.AIDecision{HRRecommendation: "hire", Confidence: 0.82},
	}
}

// TestSuccessUsesContextVacancyID covers the full happy path when the
// caller does NOT override the vacancy id: the resume's joined vacancy is
// authoritative. Asserts every step of the orchestrator runs and gets the
// data the previous step produced.
func (s *StartAnalysisSuite) TestSuccessUsesContextVacancyID() {
	t := s.T()
	ctx := t.Context()
	rc := happyResume()
	skills := []domain.VacancySkill{{Name: "go", Weight: 0.7, MustHave: true}}
	payload := happyPayload()

	s.storage.LoadResumeContextMock.Expect(ctx, "r-1", uint64(7), false).Return(rc, nil)
	s.storage.LoadVacancySkillsMock.Expect(ctx, "v-ctx").Return(skills, nil)
	s.scorer.ScoreMock.Expect(rc.ResumeText, skills).Return(payload)
	s.storage.NewIDMock.Expect().Return("a-1", nil)
	s.storage.SaveAnalysisMock.Expect(ctx, domain.SaveAnalysisInput{
		AnalysisID:     "a-1",
		VacancyID:      "v-ctx",
		CandidateID:    "c-1",
		ResumeID:       "r-1",
		VacancyVersion: 3,
		Status:         domain.StatusDone,
		Score:          payload.Score,
		Profile:        payload.Profile,
		Breakdown:      payload.Breakdown,
		AI:             payload.AI,
	}).Return(nil)

	res, err := s.svc.StartAnalysis(ctx, domain.StartAnalysisInput{
		RequestUserID: 7,
		ResumeID:      "r-1",
	})
	assert.NilError(t, err)
	assert.Equal(t, res.AnalysisID, "a-1")
	assert.Equal(t, res.Status, domain.StatusQueued)
}

// TestSuccessOverridesVacancyID covers the override branch: when the caller
// passes a non-empty input.VacancyID, that value beats the joined context's
// VacancyID. The mock expectation pins this exact id flowing through to
// LoadVacancySkills and SaveAnalysis.
func (s *StartAnalysisSuite) TestSuccessOverridesVacancyID() {
	t := s.T()
	ctx := t.Context()
	rc := happyResume() // VacancyID = "v-ctx"
	skills := []domain.VacancySkill{{Name: "go", Weight: 1}}
	payload := happyPayload()

	s.storage.LoadResumeContextMock.Expect(ctx, "r-1", uint64(7), false).Return(rc, nil)
	s.storage.LoadVacancySkillsMock.Expect(ctx, "v-override").Return(skills, nil)
	s.scorer.ScoreMock.Expect(rc.ResumeText, skills).Return(payload)
	s.storage.NewIDMock.Expect().Return("a-1", nil)
	s.storage.SaveAnalysisMock.Inspect(func(_ context.Context, in domain.SaveAnalysisInput) {
		assert.Equal(t, in.VacancyID, "v-override")
	}).Return(nil)

	res, err := s.svc.StartAnalysis(ctx, domain.StartAnalysisInput{
		RequestUserID: 7,
		ResumeID:      "r-1",
		VacancyID:     "v-override",
	})
	assert.NilError(t, err)
	assert.Equal(t, res.AnalysisID, "a-1")
}

func (s *StartAnalysisSuite) TestInvalidArgumentZeroUser() {
	t := s.T()
	res, err := s.svc.StartAnalysis(t.Context(), domain.StartAnalysisInput{ResumeID: "r-1"})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, res == nil)
}

func (s *StartAnalysisSuite) TestInvalidArgumentEmptyResume() {
	t := s.T()
	res, err := s.svc.StartAnalysis(t.Context(), domain.StartAnalysisInput{
		RequestUserID: 1,
		ResumeID:      "   ",
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, res == nil)
}

// TestNotFoundOnNilContext: when LoadResumeContext returns (nil, nil) the
// ownership filter excluded the row — orchestrator must surface as
// ErrNotFound, not nil-deref on the next step.
func (s *StartAnalysisSuite) TestNotFoundOnNilContext() {
	t := s.T()
	ctx := t.Context()

	s.storage.LoadResumeContextMock.Expect(ctx, "missing", uint64(1), false).Return(nil, nil)

	res, err := s.svc.StartAnalysis(ctx, domain.StartAnalysisInput{
		RequestUserID: 1,
		ResumeID:      "missing",
	})
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Assert(t, res == nil)
}

func (s *StartAnalysisSuite) TestLoadResumeContextError() {
	t := s.T()
	ctx := t.Context()
	storageErr := errors.New("pgx: down")

	s.storage.LoadResumeContextMock.Expect(ctx, "r-1", uint64(1), false).Return(nil, storageErr)

	res, err := s.svc.StartAnalysis(ctx, domain.StartAnalysisInput{
		RequestUserID: 1,
		ResumeID:      "r-1",
	})
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, res == nil)
}

// TestInvalidArgumentWhenVacancyUnresolved covers the corner case where the
// resume has no vacancy joined AND the caller didn't override one. The
// orchestrator must reject before touching skills/scorer/storage.
func (s *StartAnalysisSuite) TestInvalidArgumentWhenVacancyUnresolved() {
	t := s.T()
	ctx := t.Context()
	rc := happyResume()
	rc.VacancyID = "" // joined vacancy missing

	s.storage.LoadResumeContextMock.Expect(ctx, "r-1", uint64(1), false).Return(rc, nil)

	res, err := s.svc.StartAnalysis(ctx, domain.StartAnalysisInput{
		RequestUserID: 1,
		ResumeID:      "r-1",
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, res == nil)
}

func (s *StartAnalysisSuite) TestLoadVacancySkillsError() {
	t := s.T()
	ctx := t.Context()
	rc := happyResume()
	storageErr := errors.New("pgx: down")

	s.storage.LoadResumeContextMock.Expect(ctx, "r-1", uint64(7), false).Return(rc, nil)
	s.storage.LoadVacancySkillsMock.Expect(ctx, "v-ctx").Return(nil, storageErr)

	res, err := s.svc.StartAnalysis(ctx, domain.StartAnalysisInput{
		RequestUserID: 7,
		ResumeID:      "r-1",
	})
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, res == nil)
}

func (s *StartAnalysisSuite) TestNewIDError() {
	t := s.T()
	ctx := t.Context()
	rc := happyResume()
	idErr := errors.New("rand: read failed")

	s.storage.LoadResumeContextMock.Expect(ctx, "r-1", uint64(7), false).Return(rc, nil)
	s.storage.LoadVacancySkillsMock.Expect(ctx, "v-ctx").Return(nil, nil)
	s.scorer.ScoreMock.Return(happyPayload())
	s.storage.NewIDMock.Expect().Return("", idErr)

	res, err := s.svc.StartAnalysis(ctx, domain.StartAnalysisInput{
		RequestUserID: 7,
		ResumeID:      "r-1",
	})
	assert.ErrorIs(t, err, idErr)
	assert.Assert(t, res == nil)
}

func (s *StartAnalysisSuite) TestSaveAnalysisError() {
	t := s.T()
	ctx := t.Context()
	rc := happyResume()
	saveErr := errors.New("pgx: insert failed")

	s.storage.LoadResumeContextMock.Expect(ctx, "r-1", uint64(7), false).Return(rc, nil)
	s.storage.LoadVacancySkillsMock.Expect(ctx, "v-ctx").Return(nil, nil)
	s.scorer.ScoreMock.Return(happyPayload())
	s.storage.NewIDMock.Expect().Return("a-1", nil)
	s.storage.SaveAnalysisMock.Return(saveErr)

	res, err := s.svc.StartAnalysis(ctx, domain.StartAnalysisInput{
		RequestUserID: 7,
		ResumeID:      "r-1",
	})
	assert.ErrorIs(t, err, saveErr)
	assert.Assert(t, res == nil)
}

// TestUseLLMTrueButNilClientSkipsFanOut covers the defensive nil-guard:
// even if the caller asks for the LLM, a service constructed without a
// multiagent client must NOT panic and must complete the heuristic flow.
// minimock NOT expecting UpdateAIDecision proves the fan-out was skipped.
func (s *StartAnalysisSuite) TestUseLLMTrueButNilClientSkipsFanOut() {
	t := s.T()
	ctx := t.Context()
	rc := happyResume()

	s.storage.LoadResumeContextMock.Expect(ctx, "r-1", uint64(7), false).Return(rc, nil)
	s.storage.LoadVacancySkillsMock.Expect(ctx, "v-ctx").Return(nil, nil)
	s.scorer.ScoreMock.Return(happyPayload())
	s.storage.NewIDMock.Expect().Return("a-1", nil)
	s.storage.SaveAnalysisMock.Return(nil)

	res, err := s.svc.StartAnalysis(ctx, domain.StartAnalysisInput{
		RequestUserID: 7,
		ResumeID:      "r-1",
		UseLLM:        true, // requested but client is nil — fan-out must skip
	})
	assert.NilError(t, err)
	assert.Equal(t, res.AnalysisID, "a-1")
}

// TestLLMFanOutOverwritesAIDecision covers the happy fan-out path: when
// UseLLM=true and the multiagent stub returns a response, the service must
// call UpdateAIDecision with the LLM-derived AI block. The captured request
// pins the field-mapping contract.
func (s *StartAnalysisSuite) TestLLMFanOutOverwritesAIDecision() {
	t := s.T()
	ctx := t.Context()
	rc := happyResume()
	payload := happyPayload()

	s.storage.LoadResumeContextMock.Expect(ctx, "r-1", uint64(7), false).Return(rc, nil)
	s.storage.LoadVacancySkillsMock.Expect(ctx, "v-ctx").Return(nil, nil)
	s.scorer.ScoreMock.Return(payload)
	s.storage.NewIDMock.Expect().Return("a-1", nil)
	s.storage.SaveAnalysisMock.Return(nil)
	s.storage.UpdateAIDecisionMock.Inspect(func(_ context.Context, analysisID string, ai domain.AIDecision) {
		assert.Equal(t, analysisID, "a-1")
		assert.Equal(t, ai.HRRecommendation, "ma-hire")
		assert.Equal(t, ai.RawTrace, "ma-trace")
		assert.Equal(t, len(ai.AgentResults), 1)
		assert.Equal(t, ai.AgentResults[0].AgentName, "DecisionAgent")
	}).Return(nil)

	maStub := &multiagentClientStub{
		resp: &pb_multiagent.GenerateDecisionResponse{
			HrRecommendation: "ma-hire",
			Confidence:       0.91,
			HrRationale:      "ma-rationale",
			AgentResults: []*pb_multiagent.AgentResult{{
				AgentName:      "DecisionAgent",
				Summary:        "ok",
				StructuredJson: "{}",
				Confidence:     0.9,
			}},
			RawTrace: "ma-trace",
		},
	}
	svc := NewAnalysisService(s.storage, s.scorer, maStub)

	res, err := svc.StartAnalysis(ctx, domain.StartAnalysisInput{
		RequestUserID: 7,
		ResumeID:      "r-1",
		UseLLM:        true,
	})
	assert.NilError(t, err)
	assert.Equal(t, res.AnalysisID, "a-1")
	assert.Assert(t, maStub.called, "multiagent client must be invoked")
	assert.Equal(t, maStub.capturedReq.GetMatchScore(), payload.Score)
	assert.DeepEqual(t, maStub.capturedReq.GetCandidateSkills(), payload.Profile.Skills)
}

// TestLLMFanOutErrorIsSwallowed covers the fail-closed contract for LLM
// failures: heuristic decision is the authoritative fallback. The
// orchestrator returns success with the heuristic AI already saved; the
// fan-out's error MUST NOT bubble up. minimock NOT expecting
// UpdateAIDecision proves the broken response was not used.
func (s *StartAnalysisSuite) TestLLMFanOutErrorIsSwallowed() {
	t := s.T()
	ctx := t.Context()
	rc := happyResume()

	s.storage.LoadResumeContextMock.Expect(ctx, "r-1", uint64(7), false).Return(rc, nil)
	s.storage.LoadVacancySkillsMock.Expect(ctx, "v-ctx").Return(nil, nil)
	s.scorer.ScoreMock.Return(happyPayload())
	s.storage.NewIDMock.Expect().Return("a-1", nil)
	s.storage.SaveAnalysisMock.Return(nil)
	// no UpdateAIDecisionMock expectation — minimock fails the test if the
	// orchestrator calls it on the LLM error path.

	maStub := &multiagentClientStub{err: errors.New("ma: down")}
	svc := NewAnalysisService(s.storage, s.scorer, maStub)

	res, err := svc.StartAnalysis(ctx, domain.StartAnalysisInput{
		RequestUserID: 7,
		ResumeID:      "r-1",
		UseLLM:        true,
	})
	assert.NilError(t, err)
	assert.Equal(t, res.AnalysisID, "a-1")
	assert.Assert(t, maStub.called)
}

func TestStartAnalysisSuite(t *testing.T) { suite.Run(t, new(StartAnalysisSuite)) }
