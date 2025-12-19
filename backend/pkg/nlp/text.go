package nlp

import (
	"regexp"
	"strings"
)

var nonWord = regexp.MustCompile(`[^a-z0-9]+`)
var multiSpace = regexp.MustCompile(`\s+`)

// Normalize приводит строку к нижнему регистру и заменяет все "не-слова" на пробелы.
// Под "словом" понимаются a-z и 0-9 (достаточно для MVP матчей навыков).
func Normalize(s string) string {
	s = strings.ToLower(s)
	s = nonWord.ReplaceAllString(s, " ")
	s = multiSpace.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// Tokens возвращает уникальные токены нормализованного текста.
func Tokens(normalized string) map[string]struct{} {
	out := make(map[string]struct{})
	if normalized == "" {
		return out
	}
	for _, t := range strings.Split(normalized, " ") {
		if t == "" {
			continue
		}
		out[t] = struct{}{}
	}
	return out
}

// ContainsPhrase проверяет наличие фразы (уже нормализованной) как целых слов.
// Пример: "rest api" найдётся в " ... rest api ..." но не в " ... rest apis ..."
func ContainsPhrase(normalizedText, normalizedPhrase string) bool {
	if normalizedPhrase == "" {
		return false
	}
	// ensure word boundaries by padding with spaces
	hay := " " + normalizedText + " "
	needle := " " + normalizedPhrase + " "
	return strings.Contains(hay, needle)
}
