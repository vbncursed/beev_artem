package analysis_service

import (
	"context"
	"strings"

	"github.com/artem13815/hr/analysis/internal/domain"
)

func (s *AnalysisService) ListCandidatesByVacancy(ctx context.Context, in domain.ListCandidatesByVacancyInput) (*domain.ListCandidatesByVacancyResult, error) {
	if in.RequestUserID == 0 || strings.TrimSpace(in.VacancyID) == "" {
		return nil, ErrInvalidArgument
	}
	if in.Limit == 0 {
		in.Limit = 20
	}

	res, err := s.storage.ListCandidatesByVacancy(ctx, in)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrNotFound
	}
	return res, nil
}
