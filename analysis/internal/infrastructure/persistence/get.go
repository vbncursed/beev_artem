package persistence

import (
	"context"

	"github.com/artem13815/hr/analysis/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (s *AnalysisStorage) GetAnalysis(ctx context.Context, analysisID string, requestUserID uint64, isAdmin bool) (*domain.Analysis, error) {
	var a domain.Analysis
	var profileJSON []byte
	var breakdownJSON []byte
	var aiJSON []byte

	err := s.db.QueryRow(ctx, `
SELECT a.id, a.vacancy_id, a.candidate_id, a.resume_id, a.vacancy_version, a.status,
       a.match_score, a.profile_json, a.breakdown_json, a.ai_json, a.error_message,
       a.created_at, a.updated_at
FROM analyses a
JOIN candidates c ON c.id = a.candidate_id
WHERE a.id = $1 AND ($2 OR c.owner_user_id = $3)
`, analysisID, isAdmin, requestUserID).Scan(
		&a.ID,
		&a.VacancyID,
		&a.CandidateID,
		&a.ResumeID,
		&a.VacancyVersion,
		&a.Status,
		&a.MatchScore,
		&profileJSON,
		&breakdownJSON,
		&aiJSON,
		&a.ErrorMessage,
		&a.CreatedAt,
		&a.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	unmarshalJSON(profileJSON, &a.Profile)
	unmarshalJSON(breakdownJSON, &a.Breakdown)
	unmarshalJSON(aiJSON, &a.AI)

	return &a, nil
}
