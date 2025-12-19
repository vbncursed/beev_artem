package nlp

import (
	"regexp"
	"strings"
)

var (
	reNonWord = regexp.MustCompile(`[^\p{L}\p{N}]+`)
	reSpaces  = regexp.MustCompile(`\s+`)
)

// NormalizeText приводит текст к упрощённому виду для сравнения:
// - нижний регистр
// - заменяет все не-буквенно-цифровые символы на пробелы
// - схлопывает пробелы
func NormalizeText(s string) string {
	s = strings.ToLower(s)
	s = reNonWord.ReplaceAllString(s, " ")
	s = reSpaces.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// NormalizeSkill нормализует навык (фразу), чтобы корректно матчить multi-word навыки.
func NormalizeSkill(skill string) string {
	return NormalizeText(skill)
}


