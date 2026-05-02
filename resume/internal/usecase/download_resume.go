package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/artem13815/hr/resume/internal/domain"
)

// DownloadResume returns the original file bytes + filename + MIME type for
// a stored resume. Same authz contract as GetResume: bearer must own the
// candidate (or be admin). Storage handles the JOIN against candidates.
func (s *ResumeService) DownloadResume(ctx context.Context, in domain.DownloadResumeInput) (*domain.ResumeFile, error) {
	if in.RequestUserID == 0 || strings.TrimSpace(in.ResumeID) == "" {
		return nil, ErrInvalidArgument
	}

	file, err := s.storage.DownloadResume(ctx, in.ResumeID, in.RequestUserID, in.IsAdmin)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return file, nil
}
