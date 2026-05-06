package persistence

import "context"

// UpdateProfileYearsExperience overwrites profile_json.years_experience with
// the LLM-extracted value. Uses jsonb_set so we don't have to re-load and
// re-serialize the whole profile blob (skills, summary, etc.) just to bump
// one field. The third arg `true` creates the key if missing.
func (s *AnalysisStorage) UpdateProfileYearsExperience(ctx context.Context, analysisID string, yoe float32) error {
	_, err := s.db.Exec(ctx, `
UPDATE analyses
SET profile_json = jsonb_set(profile_json, '{years_experience}', to_jsonb($1::real), true),
    updated_at = NOW()
WHERE id = $2
`, yoe, analysisID)
	return err
}
