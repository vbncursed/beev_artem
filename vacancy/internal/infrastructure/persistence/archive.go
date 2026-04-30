package persistence

import (
	"context"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

func (s *VacancyStorage) ArchiveVacancy(ctx context.Context, in domain.ArchiveVacancyInput) error {
	_, err := s.db.Exec(ctx, `
UPDATE vacancies
SET status = $1,
    updated_at = NOW()
WHERE id = $2 AND ($3 OR owner_user_id = $4)
`, domain.StatusArchived, in.VacancyID, in.IsAdmin, in.OwnerUserID)
	return err
}
