package vacancy_service

import (
	"strings"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

func validateCreateInput(in domain.CreateVacancyInput) error {
	if in.OwnerUserID == 0 {
		return ErrUnauthorized
	}
	if strings.TrimSpace(in.Title) == "" {
		return ErrInvalidArgument
	}
	if len(in.Title) > 255 {
		return ErrInvalidArgument
	}
	if len(in.Description) > 4000 {
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
