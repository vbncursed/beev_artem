// Package domain holds the multiagent service's transport-independent types.
// Usecase and persistence layers MUST work in these types — never in the
// generated protobuf structs from internal/pb. The transport/grpc adapter is
// the only layer that knows about pb.
package domain

import "time"

// AgentMode mirrors pb.AgentMode but lives in the domain so we can change it
// (rename, drop a value, add a new mode) without rebuilding the whole stack.
// The integer values intentionally match the wire format so the
// transport/grpc adapter can map without a switch.
type AgentMode int

const (
	AgentModeUnspecified AgentMode = iota
	AgentModeFast
	AgentModeBalanced
	AgentModeDeep
)

// DecisionRequest is the input the heuristic decision-maker reads. Field
// names use Go naming; the transport adapter handles the snake_case wire
// translation.
type DecisionRequest struct {
	Model             string
	Mode              AgentMode
	// Role selects the prompt template (assets/prompts/<role>.txt with
	// default.txt as fallback). Free-form so adding a role is just a
	// prompt-file commit, no schema migration.
	Role              string
	CandidateSkills   []string
	MissingSkills     []string
	CandidateSummary  string
	ScoreExplanation  string
	MatchScore        float32
	VacancyMustHave   []string
	VacancyNiceToHave []string
	ResumeText        string
}

// DecisionResponse is the structured HR decision returned to analysis. The
// CreatedAt timestamp is set by the usecase at decision time, not by the
// caller.
type DecisionResponse struct {
	HRRecommendation  string
	Confidence        float32
	HRRationale       string
	CandidateFeedback string
	SoftSkillsNotes   string
	AgentResults      []AgentResult
	RawTrace          string
	CreatedAt         time.Time
}

// AgentResult is the per-agent payload (one for ExtractorAgent, one for
// DecisionAgent in the current heuristic mode).
type AgentResult struct {
	AgentName      string
	Summary        string
	StructuredJSON string
	Confidence     float32
}

// CompletionRequest is the inference call shape — provider-neutral. The
// adapter maps it onto whatever knobs the upstream API exposes. Lives
// here so that both usecase (consumer of the LLM port) and
// infrastructure/llm (producer) can depend on it without forming an
// import cycle through the generated mocks subpackage.
type CompletionRequest struct {
	// Instructions is the system-prompt-style preamble (the role-specific
	// template rendered with vacancy/candidate context).
	Instructions string
	// Input is the user-side payload — the structured data the model
	// reasons about (resume + matched/missing skills + score).
	Input string
	// Temperature in [0, 1]. Low for analytical tasks, higher for creative
	// rewrites. The HR decision pipeline pins ~0.3.
	Temperature float32
	// MaxOutputTokens caps the response length so a runaway generation
	// can't bankrupt us.
	MaxOutputTokens int
}
