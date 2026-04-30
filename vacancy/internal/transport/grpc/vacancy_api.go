package grpc

import (
	"context"

	"github.com/artem13815/hr/vacancy/internal/domain"
	"github.com/artem13815/hr/vacancy/internal/pb/vacancy_api"
)

type vacancyService interface {
	CreateVacancy(ctx context.Context, in domain.CreateVacancyInput) (*domain.Vacancy, error)
	GetVacancy(ctx context.Context, in domain.GetVacancyInput) (*domain.Vacancy, error)
	ListVacancies(ctx context.Context, in domain.ListVacanciesInput) (*domain.ListVacanciesResult, error)
	UpdateVacancy(ctx context.Context, in domain.UpdateVacancyInput) (*domain.Vacancy, error)
	ArchiveVacancy(ctx context.Context, in domain.ArchiveVacancyInput) error
}

type VacancyServiceAPI struct {
	vacancy_api.UnimplementedVacancyServiceServer
	vacancyService vacancyService
}

func NewVacancyServiceAPI(vacancyService vacancyService) *VacancyServiceAPI {
	return &VacancyServiceAPI{vacancyService: vacancyService}
}

var _ vacancy_api.VacancyServiceServer = (*VacancyServiceAPI)(nil)
