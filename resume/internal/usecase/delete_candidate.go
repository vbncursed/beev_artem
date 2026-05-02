package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/artem13815/hr/resume/internal/domain"
)

// DeleteCandidate removes a candidate together with its resume(s) via the
// FK cascade. Authorization mirrors GetCandidate: bearer must own the row
// (admin overrides). Analysis rows in the analysis service are not touched
// here — that service treats orphaned analyses as historical data.
func (s *ResumeService) DeleteCandidate(ctx context.Context, in domain.DeleteCandidateInput) error {
	if in.RequestUserID == 0 || strings.TrimSpace(in.CandidateID) == "" {
		return ErrInvalidArgument
	}

	if err := s.storage.DeleteCandidate(ctx, in.CandidateID, in.RequestUserID, in.IsAdmin); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
