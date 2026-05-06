package grpc

import (
	"context"
	"errors"
	"log/slog"

	"github.com/artem13815/hr/analysis/internal/domain"
	pb_models "github.com/artem13815/hr/analysis/internal/pb/models"
	"github.com/artem13815/hr/analysis/internal/transport/middleware"
	"github.com/artem13815/hr/analysis/internal/usecase"
	"google.golang.org/grpc/codes"
)

func (a *AnalysisServiceAPI) StartAnalysis(ctx context.Context, req *pb_models.StartAnalysisRequest) (*pb_models.StartAnalysisResponse, error) {
	userCtx, ok := middleware.Get(ctx)
	if !ok {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	slog.InfoContext(ctx, "StartAnalysis request",
		"resume_id", req.GetResumeId(),
		"vacancy_id", req.GetVacancyId(),
		"use_llm", req.GetUseLlm())
	res, err := a.analysisService.StartAnalysis(ctx, domain.StartAnalysisInput{
		RequestUserID: userCtx.UserID,
		IsAdmin:       userCtx.IsAdmin,
		ResumeID:      req.GetResumeId(),
		VacancyID:     req.GetVacancyId(),
		UseLLM:        req.GetUseLlm(),
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid analysis payload.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return &pb_models.StartAnalysisResponse{AnalysisId: res.AnalysisID, Status: toPBStatus(res.Status)}, nil
}
