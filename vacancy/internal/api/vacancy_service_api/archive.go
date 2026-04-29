package vacancy_service_api

import (
	"context"
	"errors"

	"github.com/artem13815/hr/vacancy/internal/domain"
	pb_models "github.com/artem13815/hr/vacancy/internal/pb/models"
	"github.com/artem13815/hr/vacancy/internal/services/vacancy_service"
	"google.golang.org/grpc/codes"
)

func (a *VacancyServiceAPI) ArchiveVacancy(ctx context.Context, req *pb_models.ArchiveVacancyRequest) (*pb_models.ArchiveVacancyResponse, error) {
	userCtx, err := getUserContext(ctx)
	if err != nil {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	err = a.vacancyService.ArchiveVacancy(ctx, domain.ArchiveVacancyInput{
		VacancyID:   req.GetVacancyId(),
		OwnerUserID: userCtx.UserID,
		IsAdmin:     userCtx.IsAdmin,
	})
	if err != nil {
		switch {
		case errors.Is(err, vacancy_service.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid vacancy id.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return &pb_models.ArchiveVacancyResponse{Status: "archived"}, nil
}
