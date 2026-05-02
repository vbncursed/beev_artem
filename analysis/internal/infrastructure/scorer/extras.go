package scorer

import (
	"cmp"
	"regexp"
	"slices"
	"strings"
)

// tokenRe matches identifier-like tokens used to surface "extra" skills the
// resume mentioned but the vacancy did not require. Compiled once at package
// load so each scoring call is O(text) rather than O(text + regex compile).
var tokenRe = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+.#_-]{1,}`)

// trailingPunctRe trims dangling separators we sometimes get after PDF
// reflow ("pet-\nprojects" → "pet-"). Removing them keeps extracted-skill
// chips like "pet-" out of the UI.
var trailingPunctRe = regexp.MustCompile(`[-_.]+$`)

// genericNoise filters single-word generic terms that survive token
// frequency analysis but carry no signal (e.g. "tool", "rest"). Kept short
// on purpose — over-filtering would hide real technologies.
var genericNoise = map[string]struct{}{
	"tool": {}, "tools": {}, "info": {}, "data": {},
	"team": {}, "code": {}, "task": {}, "tasks": {},
	"item": {}, "test": {}, "tests": {}, "demo": {},
	"page": {}, "site": {}, "user": {}, "users": {},
}

// extractExtraSkills mines the resume text for tokens that appear at least
// twice but were never required by the vacancy. Capped at 8 to keep the
// payload small and to avoid leaking entire CVs into the breakdown JSON.
//
// Tokens are normalized: trailing `-_.` are stripped (PDF reflow artefacts
// like "pet-" from a hyphenated line break) and a tiny stop-list filters
// the most common noise words.
func extractExtraSkills(text string, required map[string]struct{}) []string {
	all := tokenRe.FindAllString(strings.ToLower(text), -1)
	freq := map[string]int{}
	for _, raw := range all {
		t := trailingPunctRe.ReplaceAllString(raw, "")
		if len(t) < 3 {
			continue
		}
		if _, exists := required[t]; exists {
			continue
		}
		if _, noise := genericNoise[t]; noise {
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
