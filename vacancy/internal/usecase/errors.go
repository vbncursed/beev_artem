package usecase

import "errors"

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrVacancyNotFound = errors.New("vacancy not found")
	// ErrLLMUnavailable is the soft-failure signal the RoleClassifier adapter
	// returns when multiagent is down or rate-limited. resolveRole catches
	// it and falls back to the deterministic keyword detector — vacancy CRUD
	// stays available even when the LLM is not.
	ErrLLMUnavailable = errors.New("llm unavailable")
)
