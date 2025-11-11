package vacancy

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Vacancy описывает вакансию и эталонные навыки с весами.
type Vacancy struct {
	ID          uuid.UUID
	Title       string
	Description string
	CreatedAt   time.Time
	Skills      []SkillWeight
}

// SkillWeight — весовой коэффициент важности навыка в диапазоне [0,1].
type SkillWeight struct {
	Skill  string
	Weight float32
}

// Repository — порт для работы с вакансиями.
type Repository interface {
	Create(ctx context.Context, v Vacancy) error
	GetByID(ctx context.Context, id uuid.UUID) (Vacancy, error)
	List(ctx context.Context, limit, offset int) ([]Vacancy, error)
	UpdateSkills(ctx context.Context, id uuid.UUID, skills []SkillWeight) error
}


