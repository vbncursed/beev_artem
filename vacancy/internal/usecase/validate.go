package usecase

import (
	"strings"
	"unicode/utf8"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

// Length limits are enforced in **runes**, not bytes — `len(s)` on a Go
// string counts UTF-8 bytes, which double the count for Cyrillic text and
// silently rejects payloads that look fine to the user (frontend counter
// shows characters). `utf8.RuneCountInString` matches the user-visible
// glyph count and the zod `.max()` we use on the frontend.
const (
	maxTitleRunes       = 255
	maxDescriptionRunes = 4000
)

func validateCreateInput(in domain.CreateVacancyInput) error {
	if in.OwnerUserID == 0 {
		return ErrUnauthorized
	}
	if strings.TrimSpace(in.Title) == "" {
		return ErrInvalidArgument
	}
	if utf8.RuneCountInString(in.Title) > maxTitleRunes {
		return ErrInvalidArgument
	}
	if utf8.RuneCountInString(in.Description) > maxDescriptionRunes {
		return ErrInvalidArgument
	}
	if len(in.Skills) == 0 {
		return ErrInvalidArgument
	}
	for _, s := range in.Skills {
		if strings.TrimSpace(s.Name) == "" {
			return ErrInvalidArgument
		}
		if s.Weight < 0 || s.Weight > 1 {
			return ErrInvalidArgument
		}
	}
	return nil
}

func normalizeSkills(skills []domain.SkillWeight) []domain.SkillWeight {
	if len(skills) == 0 {
		return skills
	}

	hasPositive := false
	for _, s := range skills {
		if s.Weight > 0 {
			hasPositive = true
			break
		}
	}

	if hasPositive {
		return skills
	}

	equal := float32(1.0 / float32(len(skills)))
	out := make([]domain.SkillWeight, 0, len(skills))
	for _, s := range skills {
		s.Weight = equal
		out = append(out, s)
	}
	return out
}
