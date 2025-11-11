package vacancy

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// UseCase инкапсулирует приложение для работы с вакансиями.
type UseCase interface {
	Create(ctx context.Context, v Vacancy) (Vacancy, error)
	GetByID(ctx context.Context, ownerID uuid.UUID, id uuid.UUID) (Vacancy, error)
	List(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]Vacancy, error)
	UpdateSkills(ctx context.Context, ownerID uuid.UUID, id uuid.UUID, skills []SkillWeight) error
	// Admin
	GetByIDAdmin(ctx context.Context, id uuid.UUID) (Vacancy, error)
	ListAdmin(ctx context.Context, limit, offset int) ([]Vacancy, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) UseCase { return &service{repo: repo} }

func (s *service) Create(ctx context.Context, v Vacancy) (Vacancy, error) {
	v.Title = strings.TrimSpace(v.Title)
	if v.Title == "" {
		return Vacancy{}, ErrValidation("title is required")
	}
	if err := s.repo.Create(ctx, v); err != nil {
		return Vacancy{}, err
	}
	return v, nil
}

func (s *service) GetByID(ctx context.Context, ownerID uuid.UUID, id uuid.UUID) (Vacancy, error) {
	return s.repo.GetByIDForOwner(ctx, ownerID, id)
}

func (s *service) List(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]Vacancy, error) {
	return s.repo.ListByOwner(ctx, ownerID, limit, offset)
}

func (s *service) UpdateSkills(ctx context.Context, ownerID uuid.UUID, id uuid.UUID, skills []SkillWeight) error {
	return s.repo.UpdateSkillsForOwner(ctx, ownerID, id, skills)
}

// ErrValidation простая ошибка валидации.
type ErrValidation string

func (e ErrValidation) Error() string { return string(e) }

// Простой use-case без дополнительных зависимостей; валидирует вход и делегирует репозиторию.

func (s *service) GetByIDAdmin(ctx context.Context, id uuid.UUID) (Vacancy, error) {
	// fallback: админ читает без ограничений владельца
	return s.repo.GetByIDForOwner(ctx, uuid.Nil, id)
}

func (s *service) ListAdmin(ctx context.Context, limit, offset int) ([]Vacancy, error) {
	return s.repo.ListAll(ctx, limit, offset)
}
