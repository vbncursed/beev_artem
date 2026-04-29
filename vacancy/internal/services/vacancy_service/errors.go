package vacancy_service

import "errors"

var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrUnauthorized    = errors.New("unauthorized")
	ErrVacancyNotFound = errors.New("vacancy not found")
)
