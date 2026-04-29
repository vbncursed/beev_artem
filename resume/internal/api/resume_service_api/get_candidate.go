package resume_service_api

import (
	"context"
	"errors"

	"github.com/artem13815/hr/resume/internal/domain"
	pb_models "github.com/artem13815/hr/resume/internal/pb/models"
	"github.com/artem13815/hr/resume/internal/services/resume_service"
	"google.golang.org/grpc/codes"
)

func (a *ResumeServiceAPI) GetCandidate(ctx context.Context, req *pb_models.GetCandidateRequest) (*pb_models.CandidateResponse, error) {
	userCtx, err := getUserContext(ctx)
	if err != nil {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	candidate, err := a.resumeService.GetCandidate(ctx, domain.GetCandidateInput{
		RequestUserID: userCtx.UserID,
		IsAdmin:       userCtx.IsAdmin,
		CandidateID:   req.GetCandidateId(),
	})
	if err != nil {
		switch {
		case errors.Is(err, resume_service.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid candidate id.")
		case errors.Is(err, resume_service.ErrNotFound):
			return nil, newError(codes.NotFound, ErrCodeNotFound, "Candidate not found.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return &pb_models.CandidateResponse{Candidate: toPBCandidate(*candidate)}, nil
}
