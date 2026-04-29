package resume_service

import (
	"context"
	"strings"

	"github.com/artem13815/hr/resume/internal/domain"
)

func (s *ResumeService) GetResume(ctx context.Context, in domain.GetResumeInput) (*domain.Resume, error) {
	if in.RequestUserID == 0 || strings.TrimSpace(in.ResumeID) == "" {
		return nil, ErrInvalidArgument
	}

	resume, err := s.storage.GetResume(ctx, in.ResumeID, in.RequestUserID, in.IsAdmin)
	if err != nil {
		return nil, err
	}
	if resume == nil {
		return nil, ErrNotFound
	}

	return resume, nil
}
