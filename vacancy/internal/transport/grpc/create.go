package grpc

import (
	"context"
	"errors"

	"github.com/artem13815/hr/vacancy/internal/domain"
	pb_models "github.com/artem13815/hr/vacancy/internal/pb/models"
	"github.com/artem13815/hr/vacancy/internal/transport/middleware"
	"github.com/artem13815/hr/vacancy/internal/usecase"
	"google.golang.org/grpc/codes"
)

func (a *VacancyServiceAPI) CreateVacancy(ctx context.Context, req *pb_models.CreateVacancyRequest) (*pb_models.VacancyResponse, error) {
	userCtx, ok := middleware.Get(ctx)
	if !ok {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	vacancy, err := a.vacancyService.CreateVacancy(ctx, domain.CreateVacancyInput{
		OwnerUserID: userCtx.UserID,
		Title:       req.GetTitle(),
		Description: req.GetDescription(),
		Role:        req.GetRole(),
		Skills:      fromPBSkills(req.GetSkills()),
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid vacancy payload.")
		case errors.Is(err, usecase.ErrUnauthorized):
			return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return &pb_models.VacancyResponse{Vacancy: toPBVacancy(*vacancy)}, nil
}
