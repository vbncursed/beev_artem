package scorer

import (
	"math"
	"regexp"
	"strconv"
	"strings"
)

// yearsRe captures "5 лет", "7+ years", "опыт 3 года", "3-х лет", etc.
// Compiled once; we keep the bound at 1..50 years to ignore obvious noise.
// Captured group is the integer count.
var yearsRe = regexp.MustCompile(`(?i)(\d{1,2})\s*\+?\s*-?\s*[xх]?\s*(?:год|лет|year)`)

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

// round2 rounds a float32 to two decimal places — exists to keep stored
// breakdown values JSON-friendly (no trailing fp noise like 73.999998).
func round2(v float32) float32 {
	return float32(math.Round(float64(v)*100) / 100)
}

// extractYearsExperience scans for "5 лет", "7+ years", "опыт работы 3 года"
// and returns the largest plausible value. Bounded to [0, 50] to ignore
// page numbers and dates. The LLM-backed flow overrides this on success;
// heuristic mode is the only consumer at runtime.
func extractYearsExperience(text string) float32 {
	matches := yearsRe.FindAllStringSubmatch(text, -1)
	var best int
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil {
			continue
		}
		if n < 0 || n > 50 {
			continue
		}
		if n > best {
			best = n
		}
	}
	return float32(best)
}
