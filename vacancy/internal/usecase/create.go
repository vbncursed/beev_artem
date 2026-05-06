package usecase

import (
	"context"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

func (s *VacancyService) CreateVacancy(ctx context.Context, in domain.CreateVacancyInput) (*domain.Vacancy, error) {
	in.Skills = normalizeSkills(in.Skills)
	// Validate first so we don't pay for an LLM call on bad input.
	if err := validateCreateInput(in); err != nil {
		return nil, err
	}
	in.Role = s.resolveRole(ctx, in.Title, in.Description)

	return s.storage.CreateVacancy(ctx, in)
}
