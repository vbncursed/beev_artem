package domain

// ResumeContext is the joined slice of (resume, candidate, vacancy) the
// usecase needs before it can score. The persistence layer assembles it via
// a single SQL JOIN; usecase only consumes it.
type ResumeContext struct {
	ResumeID       string
	CandidateID    string
	VacancyID      string
	OwnerUserID    uint64
	FullName       string
	Email          string
	Phone          string
	ResumeText     string
	VacancyVersion uint32
	// VacancyRole is the prompt-selection key forwarded to multiagent
	// during the LLM fan-out. Empty when the vacancy has no role set —
	// multiagent handles the fallback to default prompt.
	VacancyRole string
}

// AnalysisPayload is the structured outcome of the scoring step. The Scorer
// port produces it; the usecase persists it through SaveAnalysis without
// inspecting the contents.
type AnalysisPayload struct {
	Profile   CandidateProfile
	Breakdown ScoreBreakdown
	Score     float32
	AI        AIDecision
}

// SaveAnalysisInput is the immutable record an analysis run produces. The
// usecase fills every field before handing it to the storage layer, which
// just runs an INSERT — no business logic in persistence.
type SaveAnalysisInput struct {
	AnalysisID     string
	VacancyID      string
	CandidateID    string
	ResumeID       string
	VacancyVersion uint32
	Status         string
	Score          float32
	Profile        CandidateProfile
	Breakdown      ScoreBreakdown
	AI             AIDecision
}
