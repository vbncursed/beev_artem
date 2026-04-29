package analysis_service_api

import (
	"context"
	"errors"

	"github.com/artem13815/hr/analysis/internal/domain"
	pb_models "github.com/artem13815/hr/analysis/internal/pb/models"
	"github.com/artem13815/hr/analysis/internal/services/analysis_service"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (a *AnalysisServiceAPI) ListCandidatesByVacancy(ctx context.Context, req *pb_models.ListCandidatesByVacancyRequest) (*pb_models.ListCandidatesByVacancyResponse, error) {
	userCtx, err := getUserContext(ctx)
	if err != nil {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	limit := uint32(20)
	offset := uint32(0)
	if req.GetPage() != nil {
		if req.GetPage().GetLimit() > 0 {
			limit = req.GetPage().GetLimit()
		}
		offset = req.GetPage().GetOffset()
	}

	res, err := a.analysisService.ListCandidatesByVacancy(ctx, domain.ListCandidatesByVacancyInput{
		RequestUserID:  userCtx.UserID,
		IsAdmin:        userCtx.IsAdmin,
		VacancyID:      req.GetVacancyId(),
		Limit:          limit,
		Offset:         offset,
		MinScore:       req.GetMinScore(),
		RequiredSkill:  req.GetRequiredSkill(),
		ScoreOrderDesc: toPBOrderDesc(req.GetScoreOrder()),
	})
	if err != nil {
		switch {
		case errors.Is(err, analysis_service.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid list payload.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	items := make([]*pb_models.CandidateWithAnalysis, 0, len(res.Candidates))
	for _, c := range res.Candidates {
		items = append(items, &pb_models.CandidateWithAnalysis{
			CandidateId:    c.CandidateID,
			FullName:       c.FullName,
			Email:          c.Email,
			Phone:          c.Phone,
			MatchScore:     c.MatchScore,
			AnalysisId:     c.AnalysisID,
			AnalysisStatus: toPBStatus(c.AnalysisStatus),
			CreatedAt:      timestamppb.New(c.CreatedAt),
		})
	}

	return &pb_models.ListCandidatesByVacancyResponse{Candidates: items, Page: toPBPage(limit, offset, res.Total)}, nil
}
