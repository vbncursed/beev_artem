package grpc

import (
	"context"

	pb_models "github.com/artem13815/hr/admin/internal/pb/models"
	"github.com/artem13815/hr/admin/internal/transport/middleware"
	"google.golang.org/grpc/codes"
)

func (a *AdminServiceAPI) GetOverview(ctx context.Context, _ *pb_models.GetOverviewRequest) (*pb_models.OverviewResponse, error) {
	if _, ok := middleware.Get(ctx); !ok {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	stats, err := a.svc.GetOverview(ctx)
	if err != nil {
		return nil, newError(codes.Internal, ErrCodeInternal, "Failed to fetch overview.")
	}

	return &pb_models.OverviewResponse{
		Stats: &pb_models.SystemStats{
			UsersTotal:      stats.UsersTotal,
			AdminsTotal:     stats.AdminsTotal,
			VacanciesTotal:  stats.VacanciesTotal,
			CandidatesTotal: stats.CandidatesTotal,
			AnalysesTotal:   stats.AnalysesTotal,
			AnalysesDone:    stats.AnalysesDone,
			AnalysesFailed:  stats.AnalysesFailed,
		},
	}, nil
}
