package persistence

import (
	"context"

	"github.com/artem13815/hr/analysis/internal/domain"
)

func (s *AnalysisStorage) UpdateAIDecision(ctx context.Context, analysisID string, ai domain.AIDecision) error {
	aiJSON, err := marshalJSON(ai)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(ctx, `
UPDATE analyses
SET ai_json = $1,
    updated_at = NOW()
WHERE id = $2
`, aiJSON, analysisID)
	return err
}
