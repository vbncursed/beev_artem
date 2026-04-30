package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/multiagent/internal/domain"
)

type GenerateDecisionSuite struct{ baseSuite }

// validRequest is a minimal-but-recognisable input that satisfies the
// "all-empty" guard. Tests that exercise other branches start from this and
// override only what they need to keep the assertions focused.
func validRequest() domain.DecisionRequest {
	return domain.DecisionRequest{
		Model:           "qwen-chat",
		MatchScore:      80,
		CandidateSkills: []string{"Go"},
	}
}

// TestRecommendsHire covers the top tier of the heuristic: high score with
// at most one missing skill should map to hire/0.86. We also assert the
// agent results array carries the expected 2 entries (ExtractorAgent +
// DecisionAgent) and that DecisionAgent's confidence matches the main
// confidence — the contract clients rely on.
func (s *GenerateDecisionSuite) TestRecommendsHire() {
	t := s.T()
	ctx := t.Context()
	in := validRequest()
	in.MatchScore = 80
	in.MissingSkills = []string{"k8s"} // missingCount=1 still hits the hire branch

	s.storage.StoreDecisionMock.Inspect(func(_ context.Context, req domain.DecisionRequest, _ *domain.DecisionResponse) {
		assert.Equal(t, req.MatchScore, float32(80))
	}).Return(nil)

	resp, err := s.svc.GenerateDecision(ctx, in)
	assert.NilError(t, err)
	assert.Equal(t, resp.HRRecommendation, "hire")
	assert.Equal(t, resp.Confidence, float32(0.86))
	assert.Equal(t, resp.RawTrace, "heuristic-multiagent-v1")
	assert.Assert(t, !resp.CreatedAt.IsZero())
	assert.Equal(t, len(resp.AgentResults), 2)
	assert.Equal(t, resp.AgentResults[0].AgentName, "ExtractorAgent")
	assert.Equal(t, resp.AgentResults[1].AgentName, "DecisionAgent")
	assert.Equal(t, resp.AgentResults[1].Confidence, resp.Confidence)
}

// TestRecommendsHireDropsToMaybeWhenTooMuchMissing pins the asymmetry of the
// hire tier: even with score >= 75, more than one missing skill flips the
// recommendation. Without this case a regression that loosened the
// missingCount check (e.g. <= 2) would silently inflate hire counts.
func (s *GenerateDecisionSuite) TestRecommendsHireDropsToMaybeWhenTooMuchMissing() {
	t := s.T()
	ctx := t.Context()
	in := validRequest()
	in.MatchScore = 80
	in.MissingSkills = []string{"k8s", "aws"} // missingCount=2 fails the hire branch

	s.storage.StoreDecisionMock.Return(nil)

	resp, err := s.svc.GenerateDecision(ctx, in)
	assert.NilError(t, err)
	assert.Equal(t, resp.HRRecommendation, "maybe")
	assert.Equal(t, resp.Confidence, float32(0.68))
}

func (s *GenerateDecisionSuite) TestRecommendsMaybe() {
	t := s.T()
	ctx := t.Context()
	in := validRequest()
	in.MatchScore = 50

	s.storage.StoreDecisionMock.Return(nil)

	resp, err := s.svc.GenerateDecision(ctx, in)
	assert.NilError(t, err)
	assert.Equal(t, resp.HRRecommendation, "maybe")
	assert.Equal(t, resp.Confidence, float32(0.68))
}

func (s *GenerateDecisionSuite) TestRecommendsNo() {
	t := s.T()
	ctx := t.Context()
	in := validRequest()
	in.MatchScore = 20

	s.storage.StoreDecisionMock.Return(nil)

	resp, err := s.svc.GenerateDecision(ctx, in)
	assert.NilError(t, err)
	assert.Equal(t, resp.HRRecommendation, "no")
	assert.Equal(t, resp.Confidence, float32(0.55))
}

// TestFeedbackListsMissingSkills covers the actionable-feedback branch: when
// the candidate has gaps, the feedback string MUST enumerate them so the HR
// reader knows what to ask for. A regression that lost this branch would
// surface as a generic message even when concrete data is available.
func (s *GenerateDecisionSuite) TestFeedbackListsMissingSkills() {
	t := s.T()
	ctx := t.Context()
	in := validRequest()
	in.MissingSkills = []string{"go", "sql"}

	s.storage.StoreDecisionMock.Return(nil)

	resp, err := s.svc.GenerateDecision(ctx, in)
	assert.NilError(t, err)
	assert.Assert(t, strings.Contains(resp.CandidateFeedback, "go, sql"),
		"feedback %q should list missing skills", resp.CandidateFeedback)
}

func (s *GenerateDecisionSuite) TestFeedbackDefaultWhenNoMissing() {
	t := s.T()
	ctx := t.Context()
	in := validRequest()
	in.MissingSkills = nil

	s.storage.StoreDecisionMock.Return(nil)

	resp, err := s.svc.GenerateDecision(ctx, in)
	assert.NilError(t, err)
	assert.Assert(t, strings.Contains(resp.CandidateFeedback, "Усилить резюме"),
		"feedback %q should be the default copy", resp.CandidateFeedback)
}

// TestEmptyRequestRejected covers the all-empty guard. minimock NOT
// expecting StoreDecision proves the storage was never touched: a bad
// request must fail fast at the validation gate.
func (s *GenerateDecisionSuite) TestEmptyRequestRejected() {
	t := s.T()
	resp, err := s.svc.GenerateDecision(t.Context(), domain.DecisionRequest{})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, resp == nil)
}

func (s *GenerateDecisionSuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	in := validRequest()
	storageErr := errors.New("pgx: connection refused")

	s.storage.StoreDecisionMock.Return(storageErr)

	resp, err := s.svc.GenerateDecision(ctx, in)
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, resp == nil)
}

// TestStoredResponseMatchesReturned ensures the storage row is the same
// object the caller sees back. If a future change returns a copy after
// StoreDecision succeeded, fields could drift between what the audit log
// and the caller think happened.
func (s *GenerateDecisionSuite) TestStoredResponseMatchesReturned() {
	t := s.T()
	ctx := t.Context()
	in := validRequest()
	in.MatchScore = 80
	in.MissingSkills = []string{"x"}

	var stored *domain.DecisionResponse
	s.storage.StoreDecisionMock.Inspect(func(_ context.Context, _ domain.DecisionRequest, resp *domain.DecisionResponse) {
		stored = resp
	}).Return(nil)

	resp, err := s.svc.GenerateDecision(ctx, in)
	assert.NilError(t, err)
	assert.Assert(t, stored == resp, "stored and returned must be the same pointer")
	assert.Assert(t, time.Since(resp.CreatedAt) < time.Second, "CreatedAt must be set at decision time")
}

func TestGenerateDecisionSuite(t *testing.T) { suite.Run(t, new(GenerateDecisionSuite)) }
