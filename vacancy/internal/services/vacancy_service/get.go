package vacancy_service

import (
	"context"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

func (s *VacancyService) GetVacancy(ctx context.Context, in domain.GetVacancyInput) (*domain.Vacancy, error) {
	if in.OwnerUserID == 0 || in.VacancyID == "" {
		return nil, ErrInvalidArgument
	}

	vacancy, err := s.storage.GetVacancy(ctx, in.VacancyID, in.OwnerUserID, in.IsAdmin)
	if err != nil {
		return nil, err
	}
	if vacancy == nil {
		return nil, ErrVacancyNotFound
	}

	return vacancy, nil
}
