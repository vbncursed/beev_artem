package usecase

import (
	"context"
	"log/slog"
	"time"
)

// classifyTimeout caps how long Create/Update will wait on the LLM. Set
// short on purpose: the keyword fallback is good enough that a slow
// classifier should not stall vacancy CRUD. The 5s ceiling matches the
// p99 budget for the multiagent classify call (Yandex GenerateRequest +
// rate limiter wait).
const classifyTimeout = 5 * time.Second

// resolveRole picks the multiagent prompt role for a vacancy. The LLM is
// the source of truth on the happy path; the keyword detector is the
// belt-and-suspenders fallback when the LLM is unreachable, slow, or
// returns garbage. Net effect: vacancies always land with a non-empty,
// in-vocabulary role — even when the LLM stack is degraded.
//
// We deliberately do not propagate classifier errors to the caller —
// vacancy CRUD must not fail just because the classifier hiccuped. The
// failure is observable through the slog warning and the multiagent-side
// metrics; the user sees a vacancy with a heuristic role, which can be
// re-classified later by re-saving.
func (s *VacancyService) resolveRole(ctx context.Context, title, description string) string {
	cctx, cancel := context.WithTimeout(ctx, classifyTimeout)
	defer cancel()

	role, err := s.classifier.Classify(cctx, title, description)
	if err != nil {
		slog.WarnContext(ctx, "vacancy role classifier fell back to keyword detector",
			"err", err, "title", title)
		return DetectRole(title, description)
	}
	if role == "" {
		slog.WarnContext(ctx, "vacancy role classifier returned empty role; using keyword detector",
			"title", title)
		return DetectRole(title, description)
	}
	return role
}
