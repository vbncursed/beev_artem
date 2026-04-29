package persistence

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/artem13815/hr/resume/internal/domain"
)

// CreateCandidateWithResume runs both INSERTs in one transaction so a failure
// on the resume side cannot leave a candidate row orphaned. The deferred
// Rollback is a no-op after a successful Commit.
func (s *ResumeStorage) CreateCandidateWithResume(
	ctx context.Context,
	candidateIn domain.CreateCandidateInput,
	resumeIn domain.NewResumeData,
) (*domain.Candidate, *domain.Resume, error) {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		// Rollback is idempotent; if Commit already ran it returns
		// ErrTxClosed, which we deliberately ignore.
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			// Best-effort log path is unavailable here without a logger;
			// caller will see the original err if any. Swallow.
			_ = rbErr
		}
	}()

	candidateID, err := newID()
	if err != nil {
		return nil, nil, err
	}

	var candidate domain.Candidate
	err = tx.QueryRow(ctx, `
INSERT INTO candidates (id, vacancy_id, owner_user_id, full_name, email, phone, source, comment)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, vacancy_id, owner_user_id, full_name, email, phone, source, comment, created_at
`,
		candidateID,
		candidateIn.VacancyID,
		candidateIn.RequestUserID,
		candidateIn.FullName,
		candidateIn.Email,
		candidateIn.Phone,
		candidateIn.Source,
		candidateIn.Comment,
	).Scan(
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
		return nil, nil, fmt.Errorf("insert candidate: %w", err)
	}

	resumeID, err := newID()
	if err != nil {
		return nil, nil, err
	}
	storagePath := fmt.Sprintf("db://resumes/%s", resumeID)

	var resume domain.Resume
	err = tx.QueryRow(ctx, `
INSERT INTO resumes (id, candidate_id, file_name, file_type, file_size_bytes, storage_path, extracted_text, file_data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, candidate_id, file_name, file_type, file_size_bytes, storage_path, extracted_text, created_at
`,
		resumeID,
		candidate.ID,
		resumeIn.FileName,
		resumeIn.FileType,
		len(resumeIn.Data),
		storagePath,
		resumeIn.ExtractedText,
		resumeIn.Data,
	).Scan(
		&resume.ID,
		&resume.CandidateID,
		&resume.FileName,
		&resume.FileType,
		&resume.FileSizeBytes,
		&resume.StoragePath,
		&resume.ExtractedText,
		&resume.CreatedAt,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("insert resume: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, nil, fmt.Errorf("commit: %w", err)
	}

	return &candidate, &resume, nil
}
