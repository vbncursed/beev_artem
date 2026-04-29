package vacancy_service

import (
	"context"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

func (s *VacancyService) ListVacancies(ctx context.Context, in domain.ListVacanciesInput) (*domain.ListVacanciesResult, error) {
	if in.OwnerUserID == 0 {
		return nil, ErrUnauthorized
	}
	if in.Limit == 0 {
		in.Limit = 20
	}
	if in.Limit > 100 {
		in.Limit = 100
	}

	return s.storage.ListVacancies(ctx, in)
}
