package resume_service

import (
	"context"
	"strings"

	"github.com/artem13815/hr/resume/internal/domain"
)

func (s *ResumeService) GetCandidate(ctx context.Context, in domain.GetCandidateInput) (*domain.Candidate, error) {
	if in.RequestUserID == 0 || strings.TrimSpace(in.CandidateID) == "" {
		return nil, ErrInvalidArgument
	}

	candidate, err := s.storage.GetCandidate(ctx, in.CandidateID, in.RequestUserID, in.IsAdmin)
	if err != nil {
		return nil, err
	}
	if candidate == nil {
		return nil, ErrNotFound
	}

	return candidate, nil
}
