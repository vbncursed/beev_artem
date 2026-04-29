package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/artem13815/hr/resume/internal/domain"
)

func (s *ResumeService) GetResume(ctx context.Context, in domain.GetResumeInput) (*domain.Resume, error) {
	if in.RequestUserID == 0 || strings.TrimSpace(in.ResumeID) == "" {
		return nil, ErrInvalidArgument
	}

	resume, err := s.storage.GetResume(ctx, in.ResumeID, in.RequestUserID, in.IsAdmin)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return resume, nil
}
