package persistence

import (
	"cmp"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"

	"github.com/artem13815/hr/analysis/internal/domain"
)

var tokenRe = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+.#_-]{1,}`)

func newID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func marshalJSON(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func unmarshalJSON[T any](raw []byte, out *T) {
	_ = json.Unmarshal(raw, out)
}

type resumeContext struct {
	ResumeID       string
	CandidateID    string
	VacancyID      string
	OwnerUserID    uint64
	FullName       string
	Email          string
	Phone          string
	ResumeText     string
	VacancyVersion uint32
}

func (s *AnalysisStorage) loadResumeContext(ctx context.Context, resumeID string, requestUserID uint64, isAdmin bool) (*resumeContext, error) {
	var rc resumeContext
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

func (s *AnalysisStorage) loadVacancySkills(ctx context.Context, vacancyID string) ([]domain.VacancySkill, error) {
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

func buildAnalysisPayload(resumeText string, skills []domain.VacancySkill) (domain.CandidateProfile, domain.ScoreBreakdown, float32, domain.AIDecision) {
	lowerText := strings.ToLower(resumeText)
	matched := make([]string, 0)
	missing := make([]string, 0)
	mustMissingCount := 0
	niceMatchedCount := 0
	var totalWeight float32
	var matchedWeight float32

	requiredSet := map[string]struct{}{}
	for _, sk := range skills {
		totalWeight += sk.Weight
		name := strings.TrimSpace(sk.Name)
		if name == "" {
			continue
		}
		requiredSet[strings.ToLower(name)] = struct{}{}
		if strings.Contains(lowerText, strings.ToLower(name)) {
			matched = append(matched, name)
			matchedWeight += sk.Weight
			if sk.NiceToHave {
				niceMatchedCount++
			}
		} else {
			missing = append(missing, name)
			if sk.MustHave {
				mustMissingCount++
			}
		}
	}
	if totalWeight <= 0 {
		totalWeight = 1
	}

	baseScore := (matchedWeight / totalWeight) * 100
	mustPenalty := float32(mustMissingCount) * 10
	niceBonus := float32(niceMatchedCount) * 2
	matchScore := min(max(baseScore-mustPenalty+niceBonus, 0), 100)

	extra := extractExtraSkills(lowerText, requiredSet)

	profile := domain.CandidateProfile{
		Skills:       append([]string{}, matched...),
		Summary:      summarize(resumeText),
		Technologies: append([]string{}, matched...),
	}
	breakdown := domain.ScoreBreakdown{
		MatchedSkills:   matched,
		MissingSkills:   missing,
		ExtraSkills:     extra,
		BaseScore:       round2(baseScore),
		MustHavePenalty: round2(mustPenalty),
		NiceToHaveBonus: round2(niceBonus),
		Explanation:     fmt.Sprintf("matched=%d missing=%d", len(matched), len(missing)),
	}

	recommendation := "no"
	confidence := float32(0.55)
	if matchScore >= 75 {
		recommendation = "hire"
		confidence = 0.82
	} else if matchScore >= 45 {
		recommendation = "maybe"
		confidence = 0.67
	}

	ai := domain.AIDecision{
		HRRecommendation:  recommendation,
		Confidence:        confidence,
		HRRationale:       "heuristic decision based on score and must-have gaps",
		CandidateFeedback: "improve missing skills and update portfolio",
		SoftSkillsNotes:   "not evaluated in heuristic mode",
	}

	return profile, breakdown, round2(matchScore), ai
}

func extractExtraSkills(text string, required map[string]struct{}) []string {
	all := tokenRe.FindAllString(strings.ToLower(text), -1)
	freq := map[string]int{}
	for _, t := range all {
		if len(t) < 3 {
			continue
		}
		if _, exists := required[t]; exists {
			continue
		}
		freq[t]++
	}
	type kv struct {
		k string
		v int
	}
	items := make([]kv, 0, len(freq))
	for k, v := range freq {
		if v < 2 {
			continue
		}
		items = append(items, kv{k: k, v: v})
	}
	slices.SortFunc(items, func(a, b kv) int {
		if a.v == b.v {
			return cmp.Compare(a.k, b.k)
		}
		return cmp.Compare(b.v, a.v) // descending by frequency
	})
	out := make([]string, 0, 8)
	for i := range min(len(items), 8) {
		out = append(out, items[i].k)
	}
	return out
}

func summarize(text string) string {
	text = strings.Join(strings.Fields(text), " ")
	if len(text) > 320 {
		return text[:320]
	}
	return text
}

func round2(v float32) float32 {
	return float32(math.Round(float64(v)*100) / 100)
}
