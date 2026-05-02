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
// "all-empty" guard. Tests that exercise other branches start from this
// and override only what they need to keep assertions focused.
func validRequest() domain.DecisionRequest {
	return domain.DecisionRequest{
		Model:           "qwen-chat",
		Role:            "programmer",
		MatchScore:      80,
		CandidateSkills: []string{"Go"},
		MissingSkills:   []string{"k8s"},
		ResumeText:      "5 years Go, pgx, gRPC",
	}
}

// validLLMResponse is a JSON body shaped exactly like the contract baked
// into the prompts. Tests reuse it via copy when they only need to tweak
// one field.
const validLLMResponse = `{
  "hr_recommendation": "hire",
  "confidence": 0.86,
  "hr_rationale": "Strong stack match.",
  "candidate_feedback": "Add metrics for the throughput project.",
  "soft_skills_notes": "недостаточно данных",
  "agent_results": [
    {"agent_name":"ExtractorAgent","summary":"Senior Go","structured_json":"{}","confidence":0.82},
    {"agent_name":"DecisionAgent","summary":"Stack matches","structured_json":"{}","confidence":0.86}
  ]
}`

func (s *GenerateDecisionSuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	in := validRequest()

	s.prompts.GetMock.Expect("programmer").Return("PROGRAMMER PROMPT")
	s.llm.CompleteMock.Inspect(func(_ context.Context, req domain.CompletionRequest) {
		// Instructions are the role prompt with the hard language
		// directive appended in code — both halves must reach the model.
		assert.Assert(t, strings.HasPrefix(req.Instructions, "PROGRAMMER PROMPT"),
			"instructions %q must start with role prompt", req.Instructions)
		assert.Assert(t, strings.Contains(req.Instructions, "СТРОГО на русском"),
			"instructions %q must carry the language directive", req.Instructions)
		assert.Equal(t, req.Temperature, float32(completionTemperature))
		assert.Equal(t, req.MaxOutputTokens, completionMaxTokens)
		// Input is the JSON-rendered DecisionRequest — must carry the
		// candidate skills the model will reason about.
		assert.Assert(t, strings.Contains(req.Input, `"Go"`),
			"input %q must include candidate skills", req.Input)
	}).Return(validLLMResponse, nil)
	s.storage.StoreDecisionMock.Return(nil)

	resp, err := s.svc.GenerateDecision(ctx, in)
	assert.NilError(t, err)
	assert.Equal(t, resp.HRRecommendation, "hire")
	assert.Equal(t, resp.Confidence, float32(0.86))
	assert.Equal(t, resp.RawTrace, "yandex-llm-v1")
	assert.Assert(t, !resp.CreatedAt.IsZero())
	assert.Equal(t, len(resp.AgentResults), 2)
	assert.Equal(t, resp.AgentResults[0].AgentName, "ExtractorAgent")
}

// TestEmptyRequestRejected covers the all-empty validation gate. Mocks
// NOT expecting Get / Complete / StoreDecision prove the request never
// reached prompt selection, the LLM, or the audit log.
func (s *GenerateDecisionSuite) TestEmptyRequestRejected() {
	t := s.T()
	resp, err := s.svc.GenerateDecision(t.Context(), domain.DecisionRequest{})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, resp == nil)
}

// TestPromptSelectionPassesRoleVerbatim pins the contract that the role
// flows through unchanged from request to PromptStore.Get. The store
// owns case-folding / fallback — the usecase MUST NOT pre-normalise.
func (s *GenerateDecisionSuite) TestPromptSelectionPassesRoleVerbatim() {
	t := s.T()
	in := validRequest()
	in.Role = "Accountant" // mixed case

	s.prompts.GetMock.Expect("Accountant").Return("ACCOUNTANT PROMPT")
	s.llm.CompleteMock.Return(validLLMResponse, nil)
	s.storage.StoreDecisionMock.Return(nil)

	_, err := s.svc.GenerateDecision(t.Context(), in)
	assert.NilError(t, err)
}

// TestEmptyRoleStillSelectsPrompt: even when Role is empty, the usecase
// still calls Get("") so the PromptStore can serve its default. We must
// NOT short-circuit and skip the lookup.
func (s *GenerateDecisionSuite) TestEmptyRoleStillSelectsPrompt() {
	t := s.T()
	in := validRequest()
	in.Role = ""

	s.prompts.GetMock.Expect("").Return("DEFAULT PROMPT")
	s.llm.CompleteMock.Return(validLLMResponse, nil)
	s.storage.StoreDecisionMock.Return(nil)

	_, err := s.svc.GenerateDecision(t.Context(), in)
	assert.NilError(t, err)
}

// TestLLMErrorPropagatesAndStorageNotCalled covers the soft-failure
// contract: LLM error short-circuits the pipeline; storage MUST NOT
// receive the call (no audit row for a non-decision). minimock NOT
// expecting StoreDecision proves it.
func (s *GenerateDecisionSuite) TestLLMErrorPropagatesAndStorageNotCalled() {
	t := s.T()
	in := validRequest()
	llmErr := errors.New("yandex: 503")

	s.prompts.GetMock.Return("PROMPT")
	s.llm.CompleteMock.Return("", llmErr)

	resp, err := s.svc.GenerateDecision(t.Context(), in)
	assert.ErrorIs(t, err, llmErr)
	assert.Assert(t, resp == nil)
}

// TestInvalidJSONResponseRejected covers the parser's failure path: when
// the model goes off-contract, the usecase MUST surface
// ErrLLMInvalidResponse and skip the audit row. Storage NOT expected.
func (s *GenerateDecisionSuite) TestInvalidJSONResponseRejected() {
	t := s.T()
	in := validRequest()

	s.prompts.GetMock.Return("PROMPT")
	s.llm.CompleteMock.Return("not json at all", nil)

	resp, err := s.svc.GenerateDecision(t.Context(), in)
	assert.ErrorIs(t, err, ErrLLMInvalidResponse)
	assert.Assert(t, resp == nil)
}

// TestMissingRecommendationRejected: even valid JSON, if the model omits
// hr_recommendation, the response is unusable. Pin this so a future
// schema-loosening change doesn't silently propagate empty decisions.
func (s *GenerateDecisionSuite) TestMissingRecommendationRejected() {
	t := s.T()
	in := validRequest()

	s.prompts.GetMock.Return("PROMPT")
	s.llm.CompleteMock.Return(`{"confidence": 0.9}`, nil)

	resp, err := s.svc.GenerateDecision(t.Context(), in)
	assert.ErrorIs(t, err, ErrLLMInvalidResponse)
	assert.Assert(t, resp == nil)
}

// TestStripJSONFences covers the defensive parsing for the common
// "Sure! Here is the JSON: ```json\n{...}\n```" failure mode that even
// well-prompted models sometimes produce.
func (s *GenerateDecisionSuite) TestStripJSONFences() {
	t := s.T()
	in := validRequest()
	wrapped := "Sure! Here is the JSON:\n```json\n" + validLLMResponse + "\n```"

	s.prompts.GetMock.Return("PROMPT")
	s.llm.CompleteMock.Return(wrapped, nil)
	s.storage.StoreDecisionMock.Return(nil)

	resp, err := s.svc.GenerateDecision(t.Context(), in)
	assert.NilError(t, err)
	assert.Equal(t, resp.HRRecommendation, "hire")
}

func (s *GenerateDecisionSuite) TestStorageError() {
	t := s.T()
	in := validRequest()
	storageErr := errors.New("pgx: connection refused")

	s.prompts.GetMock.Return("PROMPT")
	s.llm.CompleteMock.Return(validLLMResponse, nil)
	s.storage.StoreDecisionMock.Return(storageErr)

	resp, err := s.svc.GenerateDecision(t.Context(), in)
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, resp == nil)
}

// TestStoredResponseMatchesReturned ensures the storage row is the same
// object the caller sees back. If a future change returns a copy after
// StoreDecision succeeded, fields could drift between the audit log and
// the caller.
func (s *GenerateDecisionSuite) TestStoredResponseMatchesReturned() {
	t := s.T()
	in := validRequest()

	var stored *domain.DecisionResponse
	s.prompts.GetMock.Return("PROMPT")
	s.llm.CompleteMock.Return(validLLMResponse, nil)
	s.storage.StoreDecisionMock.Inspect(func(_ context.Context, _ domain.DecisionRequest, resp *domain.DecisionResponse) {
		stored = resp
	}).Return(nil)

	resp, err := s.svc.GenerateDecision(t.Context(), in)
	assert.NilError(t, err)
	assert.Assert(t, stored == resp, "stored and returned must be the same pointer")
	assert.Assert(t, time.Since(resp.CreatedAt) < time.Second, "CreatedAt must be set at decision time")
}

func TestGenerateDecisionSuite(t *testing.T) { suite.Run(t, new(GenerateDecisionSuite)) }
