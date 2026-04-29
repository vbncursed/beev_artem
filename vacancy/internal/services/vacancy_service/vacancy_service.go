package vacancy_service

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
	storage VacancyStorage
}

func NewVacancyService(storage VacancyStorage) *VacancyService {
	return &VacancyService{storage: storage}
}
