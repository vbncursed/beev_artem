package persistence

import (
	"context"
	"fmt"

	"github.com/artem13815/hr/admin/internal/domain"
)

// ListUsers joins auth_users with vacancy/candidate counts so the table
// can render activity metrics inline. LEFT JOIN + GROUP BY keeps users
// who never created anything (counter shows 0).
func (s *StatsStorage) ListUsers(ctx context.Context) ([]domain.AdminUserView, error) {
	const query = `
SELECT
  u.id, u.email, u.role, u.created_at,
  COALESCE(v.cnt, 0) AS vacancies_owned,
  COALESCE(c.cnt, 0) AS candidates_uploaded
FROM auth_users u
LEFT JOIN (
  SELECT owner_user_id, count(*) AS cnt FROM vacancies GROUP BY owner_user_id
) v ON v.owner_user_id = u.id
LEFT JOIN (
  SELECT owner_user_id, count(*) AS cnt FROM candidates GROUP BY owner_user_id
) c ON c.owner_user_id = u.id
ORDER BY u.id
`
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	users := make([]domain.AdminUserView, 0, 16)
	for rows.Next() {
		var u domain.AdminUserView
		if err := rows.Scan(
			&u.ID, &u.Email, &u.Role, &u.CreatedAt,
			&u.VacanciesOwned, &u.CandidatesUploaded,
		); err != nil {
			return nil, fmt.Errorf("scan user row: %w", err)
		}
		users = append(users, u)
	}
	return users, rows.Err()
}
