package vacancy_storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"github.com/artem13815/hr/vacancy/internal/domain"
	"github.com/jackc/pgx/v5"
)

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func loadSkills(ctx context.Context, q queryer, vacancyID string) ([]domain.SkillWeight, error) {
	rows, err := q.Query(ctx, `
SELECT name, weight, must_have, nice_to_have
FROM vacancy_skills
WHERE vacancy_id = $1
ORDER BY position ASC
`, vacancyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	skills := make([]domain.SkillWeight, 0)
	for rows.Next() {
		var s domain.SkillWeight
		if err := rows.Scan(&s.Name, &s.Weight, &s.MustHave, &s.NiceToHave); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return skills, nil
}

type queryer interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}
