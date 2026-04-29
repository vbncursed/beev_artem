package resume_service_api

import (
	"context"
	"errors"

	"github.com/artem13815/hr/resume/internal/domain"
	pb_models "github.com/artem13815/hr/resume/internal/pb/models"
	"github.com/artem13815/hr/resume/internal/services/resume_service"
	"google.golang.org/grpc/codes"
)

func (a *ResumeServiceAPI) CreateCandidate(ctx context.Context, req *pb_models.CreateCandidateRequest) (*pb_models.CandidateResponse, error) {
	userCtx, err := getUserContext(ctx)
	if err != nil {
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
		case errors.Is(err, resume_service.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid candidate payload.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return &pb_models.CandidateResponse{Candidate: toPBCandidate(*candidate)}, nil
}
