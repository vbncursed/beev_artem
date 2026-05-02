// Package scorer is the heuristic adapter that implements the Scorer port
// declared by usecase. It does pure computation: text + vacancy skills in,
// AnalysisPayload out. No I/O, no global state. Swap this package for an
// LLM-backed scorer later by changing only the bootstrap wiring — usecase
// stays untouched.
package scorer

import (
	"fmt"
	"strings"

	"github.com/artem13815/hr/analysis/internal/domain"
)

// HeuristicScorer implements usecase.Scorer with the keyword-matching
// algorithm. Stateless — safe to share across goroutines.
type HeuristicScorer struct{}

// New returns a fresh HeuristicScorer. Constructor exists for symmetry with
// other infrastructure adapters and to leave room for configurable thresholds.
func New() *HeuristicScorer { return &HeuristicScorer{} }

// Score runs the keyword-matching pass and returns the full AnalysisPayload.
// The algorithm:
//   - count weighted matches of each vacancy skill in the resume text;
//   - apply must-have penalty (-10 per missed must-have) and nice-to-have
//     bonus (+2 per matched nice-to-have);
//   - clamp to [0, 100];
//   - derive a recommendation tier from the final score.
func (HeuristicScorer) Score(resumeText string, skills []domain.VacancySkill) domain.AnalysisPayload {
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
		Skills:          append([]string{}, matched...),
		Summary:         summarize(resumeText),
		Technologies:    append([]string{}, matched...),
		YearsExperience: extractYearsExperience(resumeText),
	}
	breakdown := domain.ScoreBreakdown{
		MatchedSkills:   matched,
		MissingSkills:   missing,
		ExtraSkills:     extra,
		BaseScore:       round2(baseScore),
		MustHavePenalty: round2(mustPenalty),
		NiceToHaveBonus: round2(niceBonus),
		Explanation: fmt.Sprintf(
			"совпадений: %d, не хватает: %d",
			len(matched), len(missing),
		),
	}

	recommendation := "no"
	confidence := float32(0.55)
	switch {
	case matchScore >= 75:
		recommendation = "hire"
		confidence = 0.82
	case matchScore >= 45:
		recommendation = "maybe"
		confidence = 0.67
	}

	ai := domain.AIDecision{
		HRRecommendation:  recommendation,
		Confidence:        confidence,
		HRRationale:       "Эвристическая оценка по совпадению навыков и наличию обязательных требований.",
		CandidateFeedback: "Подтяните недостающие навыки и обновите портфолио — добавьте свежие проекты с релевантным стеком.",
		SoftSkillsNotes:   "Soft skills в эвристическом режиме не оценивались.",
	}

	return domain.AnalysisPayload{
		Profile:   profile,
		Breakdown: breakdown,
		Score:     round2(matchScore),
		AI:        ai,
	}
}
