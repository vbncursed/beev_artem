package persistence

import (
	"context"
	"fmt"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

func (s *VacancyStorage) CreateVacancy(ctx context.Context, in domain.CreateVacancyInput) (*domain.Vacancy, error) {
	id, err := newID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate id: %w", err)
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
INSERT INTO vacancies (id, owner_user_id, title, description, status, version)
VALUES ($1, $2, $3, $4, $5, $6)
`, id, in.OwnerUserID, in.Title, in.Description, domain.StatusActive, 1)
	if err != nil {
		return nil, err
	}

	for i, skill := range in.Skills {
		_, err = tx.Exec(ctx, `
INSERT INTO vacancy_skills (vacancy_id, position, name, weight, must_have, nice_to_have)
VALUES ($1, $2, $3, $4, $5, $6)
`, id, i, skill.Name, skill.Weight, skill.MustHave, skill.NiceToHave)
		if err != nil {
			return nil, err
		}
	}

	var vacancy domain.Vacancy
	err = tx.QueryRow(ctx, `
SELECT id, owner_user_id, title, description, status, version, created_at, updated_at
FROM vacancies
WHERE id = $1
`, id).Scan(
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
