package analysis_storage

import (
	"context"
	"fmt"
	"strings"

	"github.com/artem13815/hr/analysis/internal/domain"
)

func (s *AnalysisStorage) StartAnalysis(ctx context.Context, in domain.StartAnalysisInput) (*domain.StartAnalysisResult, error) {
	rc, err := s.loadResumeContext(ctx, in.ResumeID, in.RequestUserID, in.IsAdmin)
	if err != nil {
		return nil, err
	}

	vacancyID := rc.VacancyID
	if strings.TrimSpace(in.VacancyID) != "" {
		vacancyID = strings.TrimSpace(in.VacancyID)
	}
	if vacancyID == "" {
		return nil, fmt.Errorf("vacancy id is required")
	}

	skills, err := s.loadVacancySkills(ctx, vacancyID)
	if err != nil {
		return nil, err
	}

	profile, breakdown, score, ai := buildAnalysisPayload(rc.ResumeText, skills)

	analysisID, err := newID()
	if err != nil {
		return nil, err
	}

	profileJSON, err := marshalJSON(profile)
	if err != nil {
		return nil, err
	}
	breakdownJSON, err := marshalJSON(breakdown)
	if err != nil {
		return nil, err
	}
	aiJSON, err := marshalJSON(ai)
	if err != nil {
		return nil, err
	}

	_, err = s.db.Exec(ctx, `
INSERT INTO analyses (
  id, vacancy_id, candidate_id, resume_id, vacancy_version, status, match_score,
  profile_json, breakdown_json, ai_json, error_message
)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'')
`, analysisID, vacancyID, rc.CandidateID, rc.ResumeID, rc.VacancyVersion, domain.StatusDone, score, profileJSON, breakdownJSON, aiJSON)
	if err != nil {
		return nil, err
	}

	return &domain.StartAnalysisResult{AnalysisID: analysisID, Status: domain.StatusQueued}, nil
}
