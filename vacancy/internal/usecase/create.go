package usecase

import (
	"context"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

func (s *VacancyService) CreateVacancy(ctx context.Context, in domain.CreateVacancyInput) (*domain.Vacancy, error) {
	in.Skills = normalizeSkills(in.Skills)
	if err := validateCreateInput(in); err != nil {
		return nil, err
	}

	return s.storage.CreateVacancy(ctx, in)
}
