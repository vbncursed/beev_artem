package resume_storage

import (
	"context"

	"github.com/artem13815/hr/resume/internal/domain"
)

func (s *ResumeStorage) CreateCandidate(ctx context.Context, in domain.CreateCandidateInput) (*domain.Candidate, error) {
	id, err := newID()
	if err != nil {
		return nil, err
	}

	var candidate domain.Candidate
	err = s.db.QueryRow(ctx, `
INSERT INTO candidates (id, vacancy_id, owner_user_id, full_name, email, phone, source, comment)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, vacancy_id, owner_user_id, full_name, email, phone, source, comment, created_at
`, id, in.VacancyID, in.RequestUserID, in.FullName, in.Email, in.Phone, in.Source, in.Comment).Scan(
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
		return nil, err
	}

	return &candidate, nil
}
