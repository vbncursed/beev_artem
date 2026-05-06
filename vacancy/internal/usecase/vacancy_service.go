package usecase

//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i VacancyStorage,RoleClassifier -o ./mocks -s _mock.go -g

import (
	"context"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

type VacancyStorage interface {
	CreateVacancy(ctx context.Context, in domain.CreateVacancyInput) (*domain.Vacancy, error)
	GetVacancy(ctx context.Context, vacancyID string, ownerUserID uint64, isAdmin bool) (*domain.Vacancy, error)
	ListVacancies(ctx context.Context, in domain.ListVacanciesInput) (*domain.ListVacanciesResult, error)
	UpdateVacancy(ctx context.Context, in domain.UpdateVacancyInput) (*domain.Vacancy, error)
	ArchiveVacancy(ctx context.Context, in domain.ArchiveVacancyInput) error
}

type VacancyService struct {
	storage    VacancyStorage
	classifier RoleClassifier
}

// NewVacancyService wires the storage with the role classifier. Both are
// required — when the LLM stack is intentionally absent (e.g. some test
// harness), pass a stub classifier that returns ErrLLMUnavailable so the
// usecase falls back to the keyword detector cleanly.
func NewVacancyService(storage VacancyStorage, classifier RoleClassifier) *VacancyService {
	return &VacancyService{storage: storage, classifier: classifier}
}
