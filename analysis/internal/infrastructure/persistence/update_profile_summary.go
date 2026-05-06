package persistence

import "context"

// UpdateProfileSummary overwrites profile_json.summary with the LLM-written
// candidate blurb. Same jsonb_set approach as YOE — point edit, no full-row
// reserialize.
func (s *AnalysisStorage) UpdateProfileSummary(ctx context.Context, analysisID string, summary string) error {
	_, err := s.db.Exec(ctx, `
UPDATE analyses
SET profile_json = jsonb_set(profile_json, '{summary}', to_jsonb($1::text), true),
    updated_at = NOW()
WHERE id = $2
`, summary, analysisID)
	return err
}
