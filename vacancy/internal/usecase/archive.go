package usecase

import (
	"context"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

func (s *VacancyService) ArchiveVacancy(ctx context.Context, in domain.ArchiveVacancyInput) error {
	if in.OwnerUserID == 0 || in.VacancyID == "" {
		return ErrInvalidArgument
	}
	return s.storage.ArchiveVacancy(ctx, in)
}
