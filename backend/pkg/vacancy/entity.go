package vacancy

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Vacancy описывает вакансию и эталонные навыки с весами.
type Vacancy struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
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
	// Возвращают только данные владельца
	GetByIDForOwner(ctx context.Context, ownerID, id uuid.UUID) (Vacancy, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]Vacancy, error)
	UpdateSkillsForOwner(ctx context.Context, ownerID, id uuid.UUID, skills []SkillWeight) error
	// Админ-доступ без фильтра владельца
	GetByIDAny(ctx context.Context, id uuid.UUID) (Vacancy, error)
	ListAll(ctx context.Context, limit, offset int) ([]Vacancy, error)
	DeleteForOwner(ctx context.Context, ownerID, id uuid.UUID) error
	DeleteAny(ctx context.Context, id uuid.UUID) error
}
