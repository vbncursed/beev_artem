package vacancy_service_api

import (
	"context"
	"errors"

	"github.com/artem13815/hr/vacancy/internal/domain"
	pb_models "github.com/artem13815/hr/vacancy/internal/pb/models"
	"github.com/artem13815/hr/vacancy/internal/services/vacancy_service"
	"google.golang.org/grpc/codes"
)

func (a *VacancyServiceAPI) UpdateVacancy(ctx context.Context, req *pb_models.UpdateVacancyRequest) (*pb_models.VacancyResponse, error) {
	userCtx, err := getUserContext(ctx)
	if err != nil {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	updated, err := a.vacancyService.UpdateVacancy(ctx, domain.UpdateVacancyInput{
		VacancyID:   req.GetVacancyId(),
		OwnerUserID: userCtx.UserID,
		IsAdmin:     userCtx.IsAdmin,
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		Skills:      fromPBSkills(req.GetSkills()),
	})
	if err != nil {
		switch {
		case errors.Is(err, vacancy_service.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid vacancy payload.")
		case errors.Is(err, vacancy_service.ErrVacancyNotFound):
			return nil, newError(codes.NotFound, ErrCodeNotFound, "Vacancy not found.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return &pb_models.VacancyResponse{Vacancy: toPBVacancy(*updated)}, nil
}
