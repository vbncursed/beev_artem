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

func (a *VacancyServiceAPI) ArchiveVacancy(ctx context.Context, req *pb_models.ArchiveVacancyRequest) (*pb_models.ArchiveVacancyResponse, error) {
	userCtx, ok := middleware.Get(ctx)
	if !ok {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	err := a.vacancyService.ArchiveVacancy(ctx, domain.ArchiveVacancyInput{
		VacancyID:   req.GetVacancyId(),
		OwnerUserID: userCtx.UserID,
		IsAdmin:     userCtx.IsAdmin,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid vacancy id.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return &pb_models.ArchiveVacancyResponse{Status: "archived"}, nil
}
