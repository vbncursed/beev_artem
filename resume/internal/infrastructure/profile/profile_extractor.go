package profile

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/artem13815/hr/resume/internal/domain"
)

// Extractor satisfies usecase.ProfileExtractor.
type Extractor struct{}

func New() *Extractor { return &Extractor{} }

func (Extractor) ExtractCandidateProfile(text, fileName string) domain.CandidateProfile {
	return ExtractCandidateProfile(text, fileName)
}

// CandidateProfile remains as a local alias to keep this commit small; it
// will be removed when multiagent moves under infrastructure/profile and the
// canonical domain.CandidateProfile is the only one left.
type CandidateProfile = domain.CandidateProfile

var (
	emailRe = regexp.MustCompile(`(?i)\b[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}\b`)
	phoneRe = regexp.MustCompile(`\+?\d[\d\s()\-]{8,}\d`)
	nameRe  = regexp.MustCompile(`^[\p{L}][\p{L}\-']+(?:\s+[\p{L}][\p{L}\-']+){1,2}$`)
)

func ExtractCandidateProfile(text, fileName string) CandidateProfile {
	profile := CandidateProfile{}

	if email := emailRe.FindString(text); email != "" {
		profile.Email = strings.TrimSpace(email)
	}

	if phone := phoneRe.FindString(text); phone != "" {
		profile.Phone = normalizePhone(phone)
	}

	profile.FullName = extractNameFromText(text)
	if profile.FullName == "" {
		profile.FullName = deriveNameFromFilename(fileName)
	}

	return profile
}

func normalizePhone(raw string) string {
	raw = strings.TrimSpace(raw)
	var b strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
			continue
		}
		if r == '+' && b.Len() == 0 {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func extractNameFromText(text string) string {
	const maxLines = 40
	seen := 0
	for raw := range strings.SplitSeq(text, "\n") {
		if seen >= maxLines {
			break
		}
		seen++
		line := strings.Join(strings.Fields(raw), " ")
		if line == "" {
			continue
		}
		lower := strings.ToLower(line)
		if strings.Contains(line, "@") || strings.ContainsAny(line, "0123456789") {
			continue
		}
		if strings.Contains(lower, "github") || strings.Contains(lower, "telegram") || strings.Contains(lower, "mailto") {
			continue
		}
		if nameRe.MatchString(line) {
			return line
		}
	}
	return ""
}

func deriveNameFromFilename(fileName string) string {
	base := filepath.Base(strings.TrimSpace(fileName))
	ext := filepath.Ext(base)
	base = strings.TrimSuffix(base, ext)
	base = strings.ReplaceAll(base, "_", " ")
	base = strings.ReplaceAll(base, "-", " ")
	base = strings.Join(strings.Fields(base), " ")
	if base == "" {
		return "Candidate"
	}
	return base
}
