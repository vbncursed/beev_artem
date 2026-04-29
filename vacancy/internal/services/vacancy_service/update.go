package vacancy_service

import (
	"context"
	"strings"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

func (s *VacancyService) UpdateVacancy(ctx context.Context, in domain.UpdateVacancyInput) (*domain.Vacancy, error) {
	if in.OwnerUserID == 0 || in.VacancyID == "" {
		return nil, ErrInvalidArgument
	}
	if strings.TrimSpace(in.Title) == "" {
		return nil, ErrInvalidArgument
	}
	if len(in.Skills) == 0 {
		return nil, ErrInvalidArgument
	}

	in.Skills = normalizeSkills(in.Skills)
	updated, err := s.storage.UpdateVacancy(ctx, in)
	if err != nil {
		return nil, err
	}
	if updated == nil {
		return nil, ErrVacancyNotFound
	}

	return updated, nil
}
