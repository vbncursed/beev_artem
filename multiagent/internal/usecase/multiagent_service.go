package usecase

//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i DecisionStorage,LLM,PromptStore -o ./mocks -s _mock.go -g

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/artem13815/hr/multiagent/internal/domain"
)

// completionTemperature is intentionally low for an analytical task. The
// model has to grade a candidate against a vacancy with a fixed JSON
// schema — creativity here means hallucination.
const completionTemperature = 0.3

// completionMaxTokens caps a single decision response. 1500 is generous
// for our schema (recommendation + rationale + 2 short feedback strings +
// 2 agent_results) and keeps the bill bounded if the model misbehaves.
const completionMaxTokens = 1500

// languageDirective is appended to every role prompt as a hard,
// non-negotiable rule. We pin it in code rather than in the per-role
// templates so a new prompt file can never accidentally drop the
// constraint. The exception list keeps enum values
// (hr_recommendation: hire/maybe/no) and bare technology names readable
// — translating "hire" or "PostgreSQL" makes the audit log unparseable
// downstream.
const languageDirective = `

ВАЖНО — язык ответа:
- Все текстовые поля JSON-ответа (hr_rationale, candidate_feedback, soft_skills_notes, agent_results[*].summary) пиши СТРОГО на русском языке.
- Не используй английские слова и фразы в этих полях, кроме имён технологий, продуктов и аббревиатур (например: Go, PostgreSQL, gRPC, AWS, ИВЛ, МКБ-10).
- Поле hr_recommendation — это enum, оставь одно из значений как есть: "hire", "maybe", "no". Не переводи.
- Если в резюме встречаются английские формулировки — пересказывай их по-русски, не цитируй дословно.`

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
	llm     LLM
	prompts PromptStore
}

func NewMultiAgentService(storage DecisionStorage, llm LLM, prompts PromptStore) *MultiAgentService {
	return &MultiAgentService{storage: storage, llm: llm, prompts: prompts}
}

// GenerateDecision routes the request to a role-specific prompt, calls
// the LLM, parses the JSON response, and persists the (request, response)
// pair as an audit row.
//
// On LLM failures (provider unavailable, malformed JSON) the request is
// rejected with the underlying error so analysis upstream can fall back
// to its heuristic AI — this service does NOT carry its own fallback,
// authoritative fallback lives in analysis.
func (s *MultiAgentService) GenerateDecision(ctx context.Context, req domain.DecisionRequest) (*domain.DecisionResponse, error) {
	if isEmptyRequest(req) {
		return nil, ErrInvalidArgument
	}

	instructions := s.prompts.Get(req.Role) + languageDirective
	input := buildInput(req)

	completion, err := s.llm.Complete(ctx, domain.CompletionRequest{
		Instructions:    instructions,
		Input:           input,
		Temperature:     completionTemperature,
		MaxOutputTokens: completionMaxTokens,
	})
	if err != nil {
		return nil, err
	}

	resp, err := parseDecision(completion)
	if err != nil {
		return nil, err
	}
	resp.RawTrace = "yandex-llm-v1"
	resp.CreatedAt = time.Now()

	if err := s.storage.StoreDecision(ctx, req, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// isEmptyRequest catches accidental zero-value requests (e.g. an upstream
// bug that forgot to wire the proto -> domain conversion). The signal is
// "literally nothing useful to score on" — model and matchScore alone
// don't make a request actionable.
func isEmptyRequest(req domain.DecisionRequest) bool {
	return req.Model == "" &&
		req.MatchScore == 0 &&
		len(req.CandidateSkills) == 0 &&
		len(req.MissingSkills) == 0 &&
		req.ResumeText == ""
}

// buildInput renders the DecisionRequest into a JSON payload the LLM
// reads. JSON over plain text because the model is told to output JSON
// too — keeping both halves of the contract in the same shape minimises
// formatting drift in the response.
func buildInput(req domain.DecisionRequest) string {
	payload := map[string]any{
		"vacancy": map[string]any{
			"role":         req.Role,
			"must_have":    req.VacancyMustHave,
			"nice_to_have": req.VacancyNiceToHave,
		},
		"candidate": map[string]any{
			"skills":      req.CandidateSkills,
			"summary":     req.CandidateSummary,
			"resume_text": req.ResumeText,
		},
		"score": map[string]any{
			"match_score":       req.MatchScore,
			"missing_skills":    req.MissingSkills,
			"score_explanation": req.ScoreExplanation,
		},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		// json.Marshal of a plain map[string]any with primitive values
		// cannot fail — this path is unreachable in production. We still
		// return something usable instead of panicking.
		return fmt.Sprintf("could not marshal input: %v", err)
	}
	return string(b)
}

// llmDecision mirrors the JSON contract baked into the prompts. Field
// names are snake_case to match the JSON the model emits.
type llmDecision struct {
	HRRecommendation  string         `json:"hr_recommendation"`
	Confidence        float32        `json:"confidence"`
	HRRationale       string         `json:"hr_rationale"`
	CandidateFeedback string         `json:"candidate_feedback"`
	SoftSkillsNotes   string         `json:"soft_skills_notes"`
	AgentResults      []llmAgentItem `json:"agent_results"`
}

type llmAgentItem struct {
	AgentName      string  `json:"agent_name"`
	Summary        string  `json:"summary"`
	StructuredJSON string  `json:"structured_json"`
	Confidence     float32 `json:"confidence"`
}

// parseDecision decodes the LLM completion into a domain.DecisionResponse.
// Models occasionally wrap JSON in markdown fences or chatty preambles
// despite explicit instructions otherwise — we strip the most common
// shapes before unmarshaling. Failures map to ErrLLMInvalidResponse so
// the caller can route them as soft failures.
func parseDecision(completion string) (*domain.DecisionResponse, error) {
	raw := stripJSONFences(completion)
	var d llmDecision
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		return nil, fmt.Errorf("%w: unmarshal: %v", ErrLLMInvalidResponse, err)
	}
	if d.HRRecommendation == "" {
		return nil, fmt.Errorf("%w: missing hr_recommendation", ErrLLMInvalidResponse)
	}

	agents := make([]domain.AgentResult, 0, len(d.AgentResults))
	for _, a := range d.AgentResults {
		agents = append(agents, domain.AgentResult{
			AgentName:      a.AgentName,
			Summary:        a.Summary,
			StructuredJSON: a.StructuredJSON,
			Confidence:     a.Confidence,
		})
	}
	return &domain.DecisionResponse{
		HRRecommendation:  d.HRRecommendation,
		Confidence:        d.Confidence,
		HRRationale:       d.HRRationale,
		CandidateFeedback: d.CandidateFeedback,
		SoftSkillsNotes:   d.SoftSkillsNotes,
		AgentResults:      agents,
	}, nil
}

// stripJSONFences removes the ```json…``` wrapping models love adding,
// plus any leading/trailing whitespace. If the body starts with text
// before the first {, we also trim that — defensive parsing for the
// common "Sure! Here is the JSON: { ... }" failure mode.
func stripJSONFences(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	s = strings.TrimSpace(s)
	if i := strings.Index(s, "{"); i > 0 {
		s = s[i:]
	}
	if j := strings.LastIndex(s, "}"); j >= 0 {
		s = s[:j+1]
	}
	return s
}
