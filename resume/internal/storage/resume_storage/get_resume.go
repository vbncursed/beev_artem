package resume_storage

import (
	"context"

	"github.com/artem13815/hr/resume/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (s *ResumeStorage) GetResume(ctx context.Context, resumeID string, requestUserID uint64, isAdmin bool) (*domain.Resume, error) {
	var resume domain.Resume
	err := s.db.QueryRow(ctx, `
SELECT r.id, r.candidate_id, r.file_name, r.file_type, r.file_size_bytes, r.storage_path, r.extracted_text, r.created_at
FROM resumes r
JOIN candidates c ON c.id = r.candidate_id
WHERE r.id = $1 AND ($2 OR c.owner_user_id = $3)
`, resumeID, isAdmin, requestUserID).Scan(
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
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &resume, nil
}
