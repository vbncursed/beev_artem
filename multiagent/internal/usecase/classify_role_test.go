package usecase

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/multiagent/internal/domain"
)

type ClassifyRoleSuite struct{ baseSuite }

// stableRoles mirrors what a real PromptStore would return — sorted, no
// "default". Tests reuse it via the prompts mock.
var stableRoles = []string{"accountant", "analyst", "doctor", "electrician", "manager", "programmer"}

func validClassifyRequest() domain.RoleClassifyRequest {
	return domain.RoleClassifyRequest{
		Title:       "Senior Go Developer",
		Description: "We are looking for a backend engineer with 5 years of Go.",
	}
}

// TestSuccessProgrammer pins the happy path: model returns clean JSON with
// a known role, parser canonicalizes whitespace/case, response carries it
// through.
func (s *ClassifyRoleSuite) TestSuccessProgrammer() {
	t := s.T()
	s.prompts.ListRolesMock.Return(stableRoles)
	s.llm.CompleteMock.Inspect(func(_ context.Context, req domain.CompletionRequest) {
		// Every registered role MUST appear in the system prompt — this is
		// the contract that lets dropping a templates/<role>.txt extend
		// the classifier vocabulary without code changes.
		for _, r := range stableRoles {
			assert.Assert(t, strings.Contains(req.Instructions, r),
				"instructions must list role %q, got: %s", r, req.Instructions)
		}
		assert.Assert(t, strings.Contains(req.Instructions, "default"),
			"instructions must include the default fallback")
		assert.Equal(t, req.Temperature, float32(classifierTemperature))
		assert.Equal(t, req.MaxOutputTokens, classifierMaxTokens)
		assert.Assert(t, strings.Contains(req.Input, "Senior Go Developer"),
			"input must carry the vacancy title")
	}).Return(`{"role": "programmer"}`, nil)

	resp, err := s.svc.ClassifyRole(t.Context(), validClassifyRequest())
	assert.NilError(t, err)
	assert.Equal(t, resp.Role, "programmer")
}

// TestSuccessDefaultExplicit: the model's explicit "I'm not sure" answer
// must round-trip even though "default" is excluded from ListRoles().
func (s *ClassifyRoleSuite) TestSuccessDefaultExplicit() {
	t := s.T()
	s.prompts.ListRolesMock.Return(stableRoles)
	s.llm.CompleteMock.Return(`{"role": "default"}`, nil)

	resp, err := s.svc.ClassifyRole(t.Context(), validClassifyRequest())
	assert.NilError(t, err)
	assert.Equal(t, resp.Role, "default")
}

// TestSuccessCaseInsensitive: the parser canonicalizes "PROGRAMMER" /
// "  programmer " / "Programmer" to the same lowercase token.
func (s *ClassifyRoleSuite) TestSuccessCaseInsensitive() {
	t := s.T()
	s.prompts.ListRolesMock.Return(stableRoles)
	s.llm.CompleteMock.Return(`{"role": "  PROGRAMMER  "}`, nil)

	resp, err := s.svc.ClassifyRole(t.Context(), validClassifyRequest())
	assert.NilError(t, err)
	assert.Equal(t, resp.Role, "programmer")
}

// TestStripsMarkdownFences covers the most common LLM failure mode: the
// model wraps perfectly good JSON in ```json...``` despite the prompt.
func (s *ClassifyRoleSuite) TestStripsMarkdownFences() {
	t := s.T()
	s.prompts.ListRolesMock.Return(stableRoles)
	s.llm.CompleteMock.Return("```json\n{\"role\": \"manager\"}\n```", nil)

	resp, err := s.svc.ClassifyRole(t.Context(), validClassifyRequest())
	assert.NilError(t, err)
	assert.Equal(t, resp.Role, "manager")
}

// TestEmptyInputRejected covers the validation gate. The mocks NOT
// expecting Complete or ListRoles prove the request never reached the LLM.
func (s *ClassifyRoleSuite) TestEmptyInputRejected() {
	t := s.T()
	resp, err := s.svc.ClassifyRole(t.Context(), domain.RoleClassifyRequest{})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, resp == nil)
}

// TestWhitespaceOnlyRejected: spaces and newlines are not signal. The
// validation must trim before checking emptiness.
func (s *ClassifyRoleSuite) TestWhitespaceOnlyRejected() {
	t := s.T()
	resp, err := s.svc.ClassifyRole(t.Context(), domain.RoleClassifyRequest{
		Title:       "   ",
		Description: "\n\t",
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, resp == nil)
}

// TestUnknownRoleRejected: a model hallucinating a role outside the list
// MUST be rejected so callers can fall back. Permissive validation here
// would silently set bogus role values on vacancies.
func (s *ClassifyRoleSuite) TestUnknownRoleRejected() {
	t := s.T()
	s.prompts.ListRolesMock.Return(stableRoles)
	s.llm.CompleteMock.Return(`{"role": "astronaut"}`, nil)

	resp, err := s.svc.ClassifyRole(t.Context(), validClassifyRequest())
	assert.ErrorIs(t, err, ErrLLMInvalidResponse)
	assert.Assert(t, resp == nil)
}

// TestEmptyRoleStringRejected: even valid JSON, an empty role field is
// not a usable answer — the model must commit to something or "default".
func (s *ClassifyRoleSuite) TestEmptyRoleStringRejected() {
	t := s.T()
	s.prompts.ListRolesMock.Return(stableRoles)
	s.llm.CompleteMock.Return(`{"role": ""}`, nil)

	resp, err := s.svc.ClassifyRole(t.Context(), validClassifyRequest())
	assert.ErrorIs(t, err, ErrLLMInvalidResponse)
	assert.Assert(t, resp == nil)
}

// TestMalformedJSONRejected: a chatty model that goes fully off-contract
// must not crash the service — surfaces ErrLLMInvalidResponse.
func (s *ClassifyRoleSuite) TestMalformedJSONRejected() {
	t := s.T()
	s.prompts.ListRolesMock.Return(stableRoles)
	s.llm.CompleteMock.Return("definitely not json", nil)

	resp, err := s.svc.ClassifyRole(t.Context(), validClassifyRequest())
	assert.ErrorIs(t, err, ErrLLMInvalidResponse)
	assert.Assert(t, resp == nil)
}

// TestLLMErrorWrappedAsUnavailable: any provider-side failure (rate-limit,
// 5xx, network) gets wrapped with ErrLLMUnavailable so the vacancy adapter
// can map it to a single fallback path. Direct ErrLLMUnavailable from the
// adapter is preserved (no double-wrapping).
func (s *ClassifyRoleSuite) TestLLMErrorWrappedAsUnavailable() {
	t := s.T()
	s.prompts.ListRolesMock.Return(stableRoles)
	s.llm.CompleteMock.Return("", errors.New("yandex: 503"))

	resp, err := s.svc.ClassifyRole(t.Context(), validClassifyRequest())
	assert.ErrorIs(t, err, ErrLLMUnavailable)
	assert.Assert(t, resp == nil)
}

func (s *ClassifyRoleSuite) TestLLMSentinelPassesThrough() {
	t := s.T()
	s.prompts.ListRolesMock.Return(stableRoles)
	s.llm.CompleteMock.Return("", ErrLLMUnavailable)

	resp, err := s.svc.ClassifyRole(t.Context(), validClassifyRequest())
	assert.ErrorIs(t, err, ErrLLMUnavailable)
	assert.Assert(t, resp == nil)
}

func TestClassifyRoleSuite(t *testing.T) { suite.Run(t, new(ClassifyRoleSuite)) }
