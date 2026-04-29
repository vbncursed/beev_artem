package vacancy_service_api

import (
	"context"
	"errors"

	"github.com/artem13815/hr/vacancy/internal/domain"
	pb_models "github.com/artem13815/hr/vacancy/internal/pb/models"
	"github.com/artem13815/hr/vacancy/internal/services/vacancy_service"
	"google.golang.org/grpc/codes"
)

func (a *VacancyServiceAPI) GetVacancy(ctx context.Context, req *pb_models.GetVacancyRequest) (*pb_models.VacancyResponse, error) {
	userCtx, err := getUserContext(ctx)
	if err != nil {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	vacancy, err := a.vacancyService.GetVacancy(ctx, domain.GetVacancyInput{
		VacancyID:   req.GetVacancyId(),
		OwnerUserID: userCtx.UserID,
		IsAdmin:     userCtx.IsAdmin,
	})
	if err != nil {
		switch {
		case errors.Is(err, vacancy_service.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid vacancy id.")
		case errors.Is(err, vacancy_service.ErrVacancyNotFound):
			return nil, newError(codes.NotFound, ErrCodeNotFound, "Vacancy not found.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return &pb_models.VacancyResponse{Vacancy: toPBVacancy(*vacancy)}, nil
}
