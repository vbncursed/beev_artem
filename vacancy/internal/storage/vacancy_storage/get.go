package vacancy_storage

import (
	"context"
	"errors"

	"github.com/artem13815/hr/vacancy/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (s *VacancyStorage) GetVacancy(ctx context.Context, vacancyID string, ownerUserID uint64, isAdmin bool) (*domain.Vacancy, error) {
	var vacancy domain.Vacancy
	err := s.db.QueryRow(ctx, `
SELECT id, owner_user_id, title, description, status, version, created_at, updated_at
FROM vacancies
WHERE id = $1
  AND ($2 OR owner_user_id = $3)
`, vacancyID, isAdmin, ownerUserID).Scan(
		&vacancy.ID,
		&vacancy.OwnerUserID,
		&vacancy.Title,
		&vacancy.Description,
		&vacancy.Status,
		&vacancy.Version,
		&vacancy.CreatedAt,
		&vacancy.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	skills, err := loadSkills(ctx, s.db, vacancy.ID)
	if err != nil {
		return nil, err
	}
	vacancy.Skills = skills

	return &vacancy, nil
}
