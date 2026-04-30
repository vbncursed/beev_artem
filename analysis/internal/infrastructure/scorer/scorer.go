// Package scorer is the heuristic adapter that implements the Scorer port
// declared by usecase. It does pure computation: text + vacancy skills in,
// AnalysisPayload out. No I/O, no global state. Swap this package for an
// LLM-backed scorer later by changing only the bootstrap wiring — usecase
// stays untouched.
package scorer

import (
	"cmp"
	"fmt"
	"math"
	"regexp"
	"slices"
	"strings"

	"github.com/artem13815/hr/analysis/internal/domain"
)

// tokenRe matches identifier-like tokens used to surface "extra" skills the
// resume mentioned but the vacancy did not require. Compiled once at package
// load so each scoring call is O(text) rather than O(text + regex compile).
var tokenRe = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+.#_-]{1,}`)

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
		HRRationale:       "heuristic decision based on score and must-have gaps",
		CandidateFeedback: "improve missing skills and update portfolio",
		SoftSkillsNotes:   "not evaluated in heuristic mode",
	}

	return domain.AnalysisPayload{
		Profile:   profile,
		Breakdown: breakdown,
		Score:     round2(matchScore),
		AI:        ai,
	}
}

// extractExtraSkills mines the resume text for tokens that appear at least
// twice but were never required by the vacancy. Capped at 8 to keep the
// payload small and to avoid leaking entire CVs into the breakdown JSON.
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

// summarize collapses whitespace and truncates to a JSON-friendly preview.
// 320 chars roughly fits one tweet/elevator pitch — long enough to convey
// context, short enough to render in a list view without expansion.
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
