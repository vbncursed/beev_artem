package grpc

import (
	"context"
	"errors"

	"github.com/artem13815/hr/resume/internal/domain"
	pb_models "github.com/artem13815/hr/resume/internal/pb/models"
	"github.com/artem13815/hr/resume/internal/transport/middleware"
	"github.com/artem13815/hr/resume/internal/usecase"
	"google.golang.org/grpc/codes"
)

func (a *ResumeServiceAPI) CreateCandidate(ctx context.Context, req *pb_models.CreateCandidateRequest) (*pb_models.CandidateResponse, error) {
	userCtx, ok := middleware.Get(ctx)
	if !ok {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	candidate, err := a.resumeService.CreateCandidate(ctx, domain.CreateCandidateInput{
		RequestUserID: userCtx.UserID,
		VacancyID:     req.GetVacancyId(),
		FullName:      req.GetFullName(),
		Email:         req.GetEmail(),
		Phone:         req.GetPhone(),
		Source:        req.GetSource(),
		Comment:       req.GetComment(),
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid candidate payload.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return &pb_models.CandidateResponse{Candidate: toPBCandidate(*candidate)}, nil
}
