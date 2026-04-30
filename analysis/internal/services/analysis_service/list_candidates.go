package analysis_service

import (
	"cmp"
	"context"
	"strings"

	"github.com/artem13815/hr/analysis/internal/domain"
)

func (s *AnalysisService) ListCandidatesByVacancy(ctx context.Context, in domain.ListCandidatesByVacancyInput) (*domain.ListCandidatesByVacancyResult, error) {
	if in.RequestUserID == 0 || strings.TrimSpace(in.VacancyID) == "" {
		return nil, ErrInvalidArgument
	}
	in.Limit = cmp.Or(in.Limit, 20)

	res, err := s.storage.ListCandidatesByVacancy(ctx, in)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrNotFound
	}
	return res, nil
}
