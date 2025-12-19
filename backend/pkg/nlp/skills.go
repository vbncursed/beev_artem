package nlp

import (
	"strings"
)

// SkillVariants returns normalized variants for matching (synonyms/aliases).
// It is intentionally small for MVP; extend as needed.
func SkillVariants(skill string) []string {
	base := NormalizeSkill(skill)
	if base == "" {
		return []string{}
	}
	var out []string
	seen := map[string]struct{}{}
	add := func(s string) {
		s = NormalizeSkill(s)
		if s == "" {
			return
		}
		if _, ok := seen[s]; ok {
			return
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}

	add(base)

	// Phrase-level aliases
	switch base {
	case "postgres":
		add("postgresql")
	case "postgresql":
		add("postgres")
	case "k8s":
		add("kubernetes")
	case "kubernetes":
		add("k8s")
	case "golang":
		add("go")
	case "go":
		add("golang")
	case "js":
		add("javascript")
	case "javascript":
		add("js")
	case "ts":
		add("typescript")
	case "typescript":
		add("ts")
	case "rest":
		add("rest api")
	case "rest api":
		add("rest")
	case "ci cd":
		add("cicd")
		add("ci/cd")
	case "cicd":
		add("ci cd")
		add("ci/cd")
	}

	// Token-level expansions (for multi-word skills)
	parts := strings.Split(base, " ")
	if len(parts) > 1 {
		var expanded []string
		for _, p := range parts {
			expanded = append(expanded, TokenVariants(p)...)
		}
		if len(expanded) > 0 {
			add(strings.Join(expanded, " "))
		}
	}

	return out
}

// TokenVariants returns normalized token variants.
func TokenVariants(token string) []string {
	t := NormalizeSkill(token)
	if t == "" {
		return []string{}
	}
	switch t {
	case "postgres":
		return []string{"postgres", "postgresql"}
	case "postgresql":
		return []string{"postgresql", "postgres"}
	case "k8s":
		return []string{"k8s", "kubernetes"}
	case "kubernetes":
		return []string{"kubernetes", "k8s"}
	case "golang":
		return []string{"golang", "go"}
	case "go":
		return []string{"go", "golang"}
	case "js":
		return []string{"js", "javascript"}
	case "javascript":
		return []string{"javascript", "js"}
	case "ts":
		return []string{"ts", "typescript"}
	case "typescript":
		return []string{"typescript", "ts"}
	default:
		return []string{t}
	}
}

// Tokens splits normalized string into tokens.
func TokensList(normalized string) []string {
	if normalized == "" {
		return []string{}
	}
	return strings.Split(normalized, " ")
}
