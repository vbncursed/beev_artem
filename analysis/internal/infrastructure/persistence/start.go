package persistence

import (
	"context"
	"fmt"

	"github.com/artem13815/hr/analysis/internal/domain"
)

// LoadResumeContext returns the joined resume / candidate / vacancy slice
// the usecase needs to score. The OR-ownership clause keeps this query
// authoritative for tenant isolation: a non-admin caller can never read
// another owner's row, even if they guess the resume id.
func (s *AnalysisStorage) LoadResumeContext(ctx context.Context, resumeID string, requestUserID uint64, isAdmin bool) (*domain.ResumeContext, error) {
	var rc domain.ResumeContext
	err := s.db.QueryRow(ctx, `
SELECT r.id, r.candidate_id, c.vacancy_id, c.owner_user_id, c.full_name, c.email, c.phone, r.extracted_text, COALESCE(v.version, 1)
FROM resumes r
JOIN candidates c ON c.id = r.candidate_id
LEFT JOIN vacancies v ON v.id = c.vacancy_id
WHERE r.id = $1 AND ($2 OR c.owner_user_id = $3)
`, resumeID, isAdmin, requestUserID).Scan(
		&rc.ResumeID,
		&rc.CandidateID,
		&rc.VacancyID,
		&rc.OwnerUserID,
		&rc.FullName,
		&rc.Email,
		&rc.Phone,
		&rc.ResumeText,
		&rc.VacancyVersion,
	)
	if err != nil {
		return nil, err
	}
	return &rc, nil
}

// LoadVacancySkills reads the ordered skills set the scorer compares against.
// Weights default to 1 when the row stores 0 — a defensive normalisation
// the heuristic relies on (0 weights would zero-divide the base score).
func (s *AnalysisStorage) LoadVacancySkills(ctx context.Context, vacancyID string) ([]domain.VacancySkill, error) {
	rows, err := s.db.Query(ctx, `
SELECT name, weight, must_have, nice_to_have
FROM vacancy_skills
WHERE vacancy_id = $1
ORDER BY position ASC
`, vacancyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.VacancySkill, 0)
	for rows.Next() {
		var sk domain.VacancySkill
		if err := rows.Scan(&sk.Name, &sk.Weight, &sk.MustHave, &sk.NiceToHave); err != nil {
			return nil, err
		}
		if sk.Weight <= 0 {
			sk.Weight = 1
		}
		out = append(out, sk)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

// SaveAnalysis writes the scored result. The usecase has already:
//   - loaded the ResumeContext (so candidate/vacancy/owner are pinned),
//   - generated an analysis id (so retries are idempotent at the caller),
//   - run the scorer (so profile/breakdown/score/AI are final).
// Persistence here is a single INSERT — no business logic.
func (s *AnalysisStorage) SaveAnalysis(ctx context.Context, in domain.SaveAnalysisInput) error {
	profileJSON, err := marshalJSON(in.Profile)
	if err != nil {
		return err
	}
	breakdownJSON, err := marshalJSON(in.Breakdown)
	if err != nil {
		return err
	}
	aiJSON, err := marshalJSON(in.AI)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, `
INSERT INTO analyses (
  id, vacancy_id, candidate_id, resume_id, vacancy_version, status, match_score,
  profile_json, breakdown_json, ai_json, error_message
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'')
`,
		in.AnalysisID, in.VacancyID, in.CandidateID, in.ResumeID, in.VacancyVersion,
		in.Status, in.Score, profileJSON, breakdownJSON, aiJSON,
	)
	if err != nil {
		return fmt.Errorf("insert analysis: %w", err)
	}
	return nil
}

// NewID exposes the persistence-layer ID generator so the usecase can
// allocate an analysis id before calling SaveAnalysis. Mirroring vacancy /
// resume style where the ID is generated outside the storage tx.
func (s *AnalysisStorage) NewID() (string, error) { return newID() }
