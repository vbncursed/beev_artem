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
	ID        uuid.UUID `json:"id"`
	ResumeID  uuid.UUID `json:"resumeId"`
	VacancyID uuid.UUID `json:"vacancyId"`
	Score     float32   `json:"score"`
	Model     string    `json:"model"`
	Report    Report    `json:"report"`
	CreatedAt time.Time `json:"createdAt"`
}

// Repository — порт для сохранения/чтения анализов.
type Repository interface {
	Create(ctx context.Context, a Analysis) (Analysis, error)
	GetByID(ctx context.Context, id uuid.UUID) (Analysis, error)
	// owner/admin views
	GetByIDForOwner(ctx context.Context, ownerID, id uuid.UUID) (Analysis, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]Analysis, error)
	ListByVacancyForOwner(ctx context.Context, ownerID, vacancyID uuid.UUID, limit, offset int) ([]Analysis, error)
	ListAll(ctx context.Context, limit, offset int) ([]Analysis, error)
	ListByVacancyAny(ctx context.Context, vacancyID uuid.UUID, limit, offset int) ([]Analysis, error)
	DeleteForOwner(ctx context.Context, ownerID, id uuid.UUID) error
	DeleteAny(ctx context.Context, id uuid.UUID) error
}
