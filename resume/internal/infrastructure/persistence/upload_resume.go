package persistence

import (
	"context"
	"fmt"

	"github.com/artem13815/hr/resume/internal/domain"
)

func (s *ResumeStorage) UploadResume(ctx context.Context, in domain.UploadResumeInput) (*domain.Resume, error) {
	if _, err := s.GetCandidate(ctx, in.CandidateID, in.RequestUserID, in.IsAdmin); err != nil {
		return nil, err
	}

	id, err := newID()
	if err != nil {
		return nil, err
	}

	storagePath := fmt.Sprintf("db://resumes/%s", id)

	var resume domain.Resume
	err = s.db.QueryRow(ctx, `
INSERT INTO resumes (id, candidate_id, file_name, file_type, file_size_bytes, storage_path, extracted_text, file_data)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id, candidate_id, file_name, file_type, file_size_bytes, storage_path, extracted_text, created_at
`, id, in.CandidateID, in.FileName, in.FileType, len(in.Data), storagePath, in.ExtractedText, in.Data).Scan(
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
		return nil, err
	}

	return &resume, nil
}
