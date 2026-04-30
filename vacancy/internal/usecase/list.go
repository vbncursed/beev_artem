package usecase

import (
	"cmp"
	"context"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

func (s *VacancyService) ListVacancies(ctx context.Context, in domain.ListVacanciesInput) (*domain.ListVacanciesResult, error) {
	if in.OwnerUserID == 0 {
		return nil, ErrUnauthorized
	}
	in.Limit = min(cmp.Or(in.Limit, 20), 100)

	return s.storage.ListVacancies(ctx, in)
}
