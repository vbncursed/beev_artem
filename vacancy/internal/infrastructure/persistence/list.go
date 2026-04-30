package persistence

import (
	"context"
	"strings"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

func (s *VacancyStorage) ListVacancies(ctx context.Context, in domain.ListVacanciesInput) (*domain.ListVacanciesResult, error) {
	q := strings.TrimSpace(in.Query)
	pattern := "%" + q + "%"

	var total uint64
	err := s.db.QueryRow(ctx, `
SELECT COUNT(*)
FROM vacancies
WHERE ($1 OR owner_user_id = $2)
  AND ($3 = '%%' OR title ILIKE $3 OR description ILIKE $3)
`, in.IsAdmin, in.OwnerUserID, pattern).Scan(&total)
	if err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
SELECT id, owner_user_id, title, description, status, version, created_at, updated_at
FROM vacancies
WHERE ($1 OR owner_user_id = $2)
  AND ($3 = '%%' OR title ILIKE $3 OR description ILIKE $3)
ORDER BY created_at DESC
LIMIT $4 OFFSET $5
`, in.IsAdmin, in.OwnerUserID, pattern, in.Limit, in.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	vacancies := make([]domain.Vacancy, 0)
	for rows.Next() {
		var v domain.Vacancy
		if err := rows.Scan(&v.ID, &v.OwnerUserID, &v.Title, &v.Description, &v.Status, &v.Version, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		skills, err := loadSkills(ctx, s.db, v.ID)
		if err != nil {
			return nil, err
		}
		v.Skills = skills
		vacancies = append(vacancies, v)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return &domain.ListVacanciesResult{Vacancies: vacancies, Total: total}, nil
}
