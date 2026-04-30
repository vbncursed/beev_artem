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
