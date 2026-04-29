package vacancy_service_api

import (
	"cmp"
	"context"
	"errors"

	"github.com/artem13815/hr/vacancy/internal/domain"
	pb_models "github.com/artem13815/hr/vacancy/internal/pb/models"
	"github.com/artem13815/hr/vacancy/internal/services/vacancy_service"
	"google.golang.org/grpc/codes"
)

func (a *VacancyServiceAPI) ListVacancies(ctx context.Context, req *pb_models.ListVacanciesRequest) (*pb_models.ListVacanciesResponse, error) {
	userCtx, err := getUserContext(ctx)
	if err != nil {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	var limit, offset uint32 = 20, 0
	if p := req.GetPage(); p != nil {
		limit = cmp.Or(p.GetLimit(), limit)
		offset = p.GetOffset()
	}

	res, err := a.vacancyService.ListVacancies(ctx, domain.ListVacanciesInput{
		OwnerUserID: userCtx.UserID,
		IsAdmin:     userCtx.IsAdmin,
		Limit:       limit,
		Offset:      offset,
		Query:       req.GetQuery(),
	})
	if err != nil {
		switch {
		case errors.Is(err, vacancy_service.ErrUnauthorized):
			return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	out := make([]*pb_models.Vacancy, 0, len(res.Vacancies))
	for _, v := range res.Vacancies {
		out = append(out, toPBVacancy(v))
	}

	return &pb_models.ListVacanciesResponse{
		Vacancies: out,
		Page:      toPBPage(limit, offset, res.Total),
	}, nil
}
