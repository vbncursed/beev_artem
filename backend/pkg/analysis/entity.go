package analysis

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Report — результирующий отчёт анализа.
type Report struct {
	CandidateSummary             string   `json:"candidateSummary"`
	MatchedSkills                []string `json:"matchedSkills"`
	MissingSkills                []string `json:"missingSkills"`
	UniqueStrengths              []string `json:"uniqueStrengths"`
	AIRecommendationForHR        string   `json:"aiRecommendationForHR"`
	AIRecommendationForCandidate []string `json:"aiRecommendationForCandidate"`
}

// Analysis хранит связи и численный скор.
type Analysis struct {
	ID        uuid.UUID
	ResumeID  uuid.UUID
	VacancyID uuid.UUID
	Score     float32
	Model     string
	Report    Report
	CreatedAt time.Time
}

// Repository — порт для сохранения/чтения анализов.
type Repository interface {
	Create(ctx context.Context, a Analysis) error
	GetByID(ctx context.Context, id uuid.UUID) (Analysis, error)
}
