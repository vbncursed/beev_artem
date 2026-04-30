package usecase

import (
	"context"
	"strings"

	"github.com/artem13815/hr/analysis/internal/domain"
)

func (s *AnalysisService) GetAnalysis(ctx context.Context, in domain.GetAnalysisInput) (*domain.Analysis, error) {
	if in.RequestUserID == 0 || strings.TrimSpace(in.AnalysisID) == "" {
		return nil, ErrInvalidArgument
	}

	res, err := s.storage.GetAnalysis(ctx, in.AnalysisID, in.RequestUserID, in.IsAdmin)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, ErrNotFound
	}
	return res, nil
}
