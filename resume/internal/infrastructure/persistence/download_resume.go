package persistence

import (
	"context"
	"errors"

	"github.com/artem13815/hr/resume/internal/domain"
	"github.com/jackc/pgx/v5"
)

// DownloadResume fetches the binary payload + minimal metadata for a single
// resume. Authorization mirrors GetResume — only the owner (or admin) can
// pull the file bytes. We deliberately keep this separate from GetResume so
// the lighter listing path doesn't drag MB-sized BYTEA columns through pgx.
func (s *ResumeStorage) DownloadResume(
	ctx context.Context,
	resumeID string,
	requestUserID uint64,
	isAdmin bool,
) (*domain.ResumeFile, error) {
	var file domain.ResumeFile
	err := s.db.QueryRow(ctx, `
SELECT r.file_name, r.file_type, r.file_data
FROM resumes r
JOIN candidates c ON c.id = r.candidate_id
WHERE r.id = $1 AND ($2 OR c.owner_user_id = $3)
`, resumeID, isAdmin, requestUserID).Scan(
		&file.FileName,
		&file.FileType,
		&file.Data,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &file, nil
}
