package domain

import "time"

const (
	StatusQueued  = "queued"
	StatusRunning = "running"
	StatusDone    = "done"
	StatusFailed  = "failed"
)

type CandidateProfile struct {
	Skills          []string `json:"skills"`
	YearsExperience float32  `json:"years_experience"`
	Positions       []string `json:"positions"`
	Technologies    []string `json:"technologies"`
	Education       []string `json:"education"`
	Summary         string   `json:"summary"`
}

type ScoreBreakdown struct {
	MatchedSkills   []string `json:"matched_skills"`
	MissingSkills   []string `json:"missing_skills"`
	ExtraSkills     []string `json:"extra_skills"`
	BaseScore       float32  `json:"base_score"`
	MustHavePenalty float32  `json:"must_have_penalty"`
	NiceToHaveBonus float32  `json:"nice_to_have_bonus"`
	Explanation     string   `json:"explanation"`
}

type AgentResult struct {
	AgentName      string  `json:"agent_name"`
	Summary        string  `json:"summary"`
	StructuredJSON string  `json:"structured_json"`
	Confidence     float32 `json:"confidence"`
}

type AIDecision struct {
	HRRecommendation  string        `json:"hr_recommendation"`
	Confidence        float32       `json:"confidence"`
	HRRationale       string        `json:"hr_rationale"`
	CandidateFeedback string        `json:"candidate_feedback"`
	SoftSkillsNotes   string        `json:"soft_skills_notes"`
	AgentResults      []AgentResult `json:"agent_results"`
	RawTrace          string        `json:"raw_trace"`
}

type Analysis struct {
	ID             string
	VacancyID      string
	CandidateID    string
	ResumeID       string
	VacancyVersion uint32
	Status         string
	MatchScore     float32
	Profile        CandidateProfile
	Breakdown      ScoreBreakdown
	AI             AIDecision
	ErrorMessage   string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type StartAnalysisInput struct {
	RequestUserID uint64
	IsAdmin       bool
	ResumeID      string
	VacancyID     string
	UseLLM        bool
}

type StartAnalysisResult struct {
	AnalysisID string
	Status     string
}

type GetAnalysisInput struct {
	RequestUserID uint64
	IsAdmin       bool
	AnalysisID    string
}

type ListCandidatesByVacancyInput struct {
	RequestUserID  uint64
	IsAdmin        bool
	VacancyID      string
	Limit          uint32
	Offset         uint32
	MinScore       float32
	RequiredSkill  string
	ScoreOrderDesc bool
}

type CandidateWithAnalysis struct {
	CandidateID    string
	FullName       string
	Email          string
	Phone          string
	MatchScore     float32
	AnalysisID     string
	AnalysisStatus string
	CreatedAt      time.Time
}

type ListCandidatesByVacancyResult struct {
	Candidates []CandidateWithAnalysis
	Total      uint64
}

type VacancySkill struct {
	Name       string
	Weight     float32
	MustHave   bool
	NiceToHave bool
}
