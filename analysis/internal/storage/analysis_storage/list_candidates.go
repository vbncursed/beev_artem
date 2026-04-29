package analysis_storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/artem13815/hr/analysis/internal/domain"
)

func (s *AnalysisStorage) ListCandidatesByVacancy(ctx context.Context, in domain.ListCandidatesByVacancyInput) (*domain.ListCandidatesByVacancyResult, error) {
	var access int
	err := s.db.QueryRow(ctx, `
SELECT 1
FROM vacancies
WHERE id = $1 AND ($2 OR owner_user_id = $3)
`, in.VacancyID, in.IsAdmin, in.RequestUserID).Scan(&access)
	if err != nil {
		return nil, err
	}

	requiredSkill := strings.ToLower(strings.TrimSpace(in.RequiredSkill))
	order := "DESC"
	if !in.ScoreOrderDesc {
		order = "ASC"
	}

	countQuery := `
WITH latest AS (
  SELECT DISTINCT ON (a.candidate_id)
    a.id, a.candidate_id, a.match_score, a.breakdown_json
  FROM analyses a
  WHERE a.vacancy_id = $1
  ORDER BY a.candidate_id, a.created_at DESC
)
SELECT COUNT(*)
FROM latest
WHERE ($2::real <= 0 OR latest.match_score >= $2)
  AND ($3 = '' OR (latest.breakdown_json->'matched_skills') ? $3)
`
	var total uint64
	if err := s.db.QueryRow(ctx, countQuery, in.VacancyID, in.MinScore, requiredSkill).Scan(&total); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(`
WITH latest AS (
  SELECT DISTINCT ON (a.candidate_id)
    a.id, a.candidate_id, a.status, a.match_score, a.created_at
  FROM analyses a
  WHERE a.vacancy_id = $1
  ORDER BY a.candidate_id, a.created_at DESC
)
SELECT latest.candidate_id, c.full_name, c.email, c.phone,
       latest.match_score, latest.id, latest.status, latest.created_at
FROM latest
JOIN candidates c ON c.id = latest.candidate_id
LEFT JOIN analyses ax ON ax.id = latest.id
WHERE ($2::real <= 0 OR latest.match_score >= $2)
  AND ($3 = '' OR (ax.breakdown_json->'matched_skills') ? $3)
ORDER BY latest.match_score %s, latest.created_at DESC
LIMIT $4 OFFSET $5
`, order)

	rows, err := s.db.Query(ctx, query, in.VacancyID, in.MinScore, requiredSkill, in.Limit, in.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.CandidateWithAnalysis, 0)
	for rows.Next() {
		var item domain.CandidateWithAnalysis
		if err := rows.Scan(
			&item.CandidateID,
			&item.FullName,
			&item.Email,
			&item.Phone,
			&item.MatchScore,
			&item.AnalysisID,
			&item.AnalysisStatus,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	return &domain.ListCandidatesByVacancyResult{Candidates: out, Total: total}, nil
}
