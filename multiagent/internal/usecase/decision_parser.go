package usecase

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/artem13815/hr/multiagent/internal/domain"
)

// llmDecision is the wire-shape of the LLM's JSON output. Decoupled from
// `domain.DecisionResponse` so a flaky model that adds extra fields or
// renames keys can't break our internal contract.
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

// parseDecision turns the model's text completion into a domain response.
// We strip markdown fences first because chatty models love wrapping JSON
// in ```json...``` even when prompted not to.
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

// stripJSONFences trims markdown code fences and dangling prose around a
// JSON object. Defensive against models that emit "Here is the JSON: ...".
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
