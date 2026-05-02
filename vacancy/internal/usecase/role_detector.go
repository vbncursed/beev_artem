package usecase

import "strings"

// roleKeywords maps a multiagent prompt role to keyword fragments that signal
// it. Order matters — the first role with a hit wins. Specific roles come
// before generic ones so a "data analyst" vacancy is not eaten by the
// programmer catch-all. Keywords are lowercase substrings; the haystack is
// lowercased before matching, so case is irrelevant.
//
// The list mirrors the prompt templates baked into multiagent
// (assets/prompts/<role>.txt). Adding a new role: drop a template there and
// extend this table.
var roleKeywords = []struct {
	role     string
	keywords []string
}{
	{"accountant", []string{
		"бухгалт", "accountant", "финансист", "финанс. директор",
		"главбух", "аудитор", "auditor",
	}},
	{"doctor", []string{
		"врач", "doctor", "physician", "медиц", "терапевт",
		"хирург", "кардиолог", "педиатр", "стоматолог",
	}},
	{"electrician", []string{
		"электрик", "electrician", "электромонт",
	}},
	{"analyst", []string{
		"аналитик", "analyst", "analytics",
		"data scientist", "data engineer", "ml engineer",
		"бизнес-аналит", "системный аналит",
	}},
	{"manager", []string{
		"менедж", "manager", "руководит", "тимлид", "team lead",
		"тим-лид", "product owner", "проджект", "project manager",
		"директор", "director", "head of", "scrum master",
	}},
	{"programmer", []string{
		"програм", "разработ", "developer", " dev ", "engineer",
		"frontend", "backend", "fullstack", "full-stack", "full stack",
		"devops", "sre", "qa", "тестировщ",
		"mobile", "мобильн", "ios", "android",
		"golang", " go ", "python", "java", "javascript", "typescript",
		"ruby", "php", "rust", "c++", "c#", "scala", "kotlin", "swift",
		"node.js", "react", "vue", "angular",
	}},
}

// DetectRole infers the multiagent prompt role from a vacancy's title and
// description. Falls back to "default" when no keyword matches; multiagent
// also falls back to default.txt for unknown roles, so this is a belt-and-
// suspenders contract.
func DetectRole(title, description string) string {
	haystack := " " + strings.ToLower(title+" "+description) + " "
	for _, r := range roleKeywords {
		for _, kw := range r.keywords {
			if strings.Contains(haystack, kw) {
				return r.role
			}
		}
	}
	return "default"
}
