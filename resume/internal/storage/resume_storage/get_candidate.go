package resume_storage

import (
	"context"

	"github.com/artem13815/hr/resume/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (s *ResumeStorage) GetCandidate(ctx context.Context, candidateID string, requestUserID uint64, isAdmin bool) (*domain.Candidate, error) {
	var candidate domain.Candidate
	err := s.db.QueryRow(ctx, `
SELECT id, vacancy_id, owner_user_id, full_name, email, phone, source, comment, created_at
FROM candidates
WHERE id = $1 AND ($2 OR owner_user_id = $3)
`, candidateID, isAdmin, requestUserID).Scan(
		&candidate.ID,
		&candidate.VacancyID,
		&candidate.OwnerUserID,
		&candidate.FullName,
		&candidate.Email,
		&candidate.Phone,
		&candidate.Source,
		&candidate.Comment,
		&candidate.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &candidate, nil
}
