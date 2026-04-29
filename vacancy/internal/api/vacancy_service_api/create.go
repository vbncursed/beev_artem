package vacancy_service_api

import (
	"context"
	"errors"

	"github.com/artem13815/hr/vacancy/internal/domain"
	pb_models "github.com/artem13815/hr/vacancy/internal/pb/models"
	"github.com/artem13815/hr/vacancy/internal/services/vacancy_service"
	"google.golang.org/grpc/codes"
)

func (a *VacancyServiceAPI) CreateVacancy(ctx context.Context, req *pb_models.CreateVacancyRequest) (*pb_models.VacancyResponse, error) {
	userCtx, err := getUserContext(ctx)
	if err != nil {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	vacancy, err := a.vacancyService.CreateVacancy(ctx, domain.CreateVacancyInput{
		OwnerUserID: userCtx.UserID,
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		Skills:      fromPBSkills(req.GetSkills()),
	})
	if err != nil {
		switch {
		case errors.Is(err, vacancy_service.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid vacancy payload.")
		case errors.Is(err, vacancy_service.ErrUnauthorized):
			return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return &pb_models.VacancyResponse{Vacancy: toPBVacancy(*vacancy)}, nil
}
