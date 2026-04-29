package vacancy_storage

import (
	"context"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

func (s *VacancyStorage) UpdateVacancy(ctx context.Context, in domain.UpdateVacancyInput) (*domain.Vacancy, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	cmd, err := tx.Exec(ctx, `
UPDATE vacancies
SET title = $1,
    description = $2,
    version = version + 1,
    updated_at = NOW()
WHERE id = $3 AND ($4 OR owner_user_id = $5)
`, in.Title, in.Description, in.VacancyID, in.IsAdmin, in.OwnerUserID)
	if err != nil {
		return nil, err
	}
	if cmd.RowsAffected() == 0 {
		return nil, nil
	}

	_, err = tx.Exec(ctx, `DELETE FROM vacancy_skills WHERE vacancy_id = $1`, in.VacancyID)
	if err != nil {
		return nil, err
	}

	for i, skill := range in.Skills {
		_, err = tx.Exec(ctx, `
INSERT INTO vacancy_skills (vacancy_id, position, name, weight, must_have, nice_to_have)
VALUES ($1, $2, $3, $4, $5, $6)
`, in.VacancyID, i, skill.Name, skill.Weight, skill.MustHave, skill.NiceToHave)
		if err != nil {
			return nil, err
		}
	}

	var vacancy domain.Vacancy
	err = tx.QueryRow(ctx, `
SELECT id, owner_user_id, title, description, status, version, created_at, updated_at
FROM vacancies
WHERE id = $1
`, in.VacancyID).Scan(
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
		return nil, err
	}

	vacancy.Skills = in.Skills

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &vacancy, nil
}
