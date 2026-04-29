package usecase

import (
	"context"
	"strings"

	"github.com/artem13815/hr/resume/internal/domain"
)

func (s *ResumeService) CreateCandidate(ctx context.Context, in domain.CreateCandidateInput) (*domain.Candidate, error) {
	if in.RequestUserID == 0 || strings.TrimSpace(in.VacancyID) == "" || strings.TrimSpace(in.FullName) == "" {
		return nil, ErrInvalidArgument
	}

	candidate, err := s.storage.CreateCandidate(ctx, in)
	if err != nil {
		return nil, err
	}
	return candidate, nil
}
