package grpc

import (
	"context"
	"errors"

	"github.com/artem13815/hr/analysis/internal/domain"
	pb_models "github.com/artem13815/hr/analysis/internal/pb/models"
	"github.com/artem13815/hr/analysis/internal/transport/middleware"
	"github.com/artem13815/hr/analysis/internal/usecase"
	"google.golang.org/grpc/codes"
)

func (a *AnalysisServiceAPI) GetAnalysis(ctx context.Context, req *pb_models.GetAnalysisRequest) (*pb_models.AnalysisResponse, error) {
	userCtx, ok := middleware.Get(ctx)
	if !ok {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	analysis, err := a.analysisService.GetAnalysis(ctx, domain.GetAnalysisInput{
		RequestUserID: userCtx.UserID,
		IsAdmin:       userCtx.IsAdmin,
		AnalysisID:    req.GetAnalysisId(),
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid analysis id.")
		case errors.Is(err, usecase.ErrNotFound):
			return nil, newError(codes.NotFound, ErrCodeNotFound, "Analysis not found.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return &pb_models.AnalysisResponse{Analysis: toPBAnalysis(*analysis)}, nil
}
