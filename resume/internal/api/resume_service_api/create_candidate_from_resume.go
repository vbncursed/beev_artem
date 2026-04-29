package resume_service_api

import (
	"context"
	"errors"

	"github.com/artem13815/hr/resume/internal/domain"
	pb_models "github.com/artem13815/hr/resume/internal/pb/models"
	"github.com/artem13815/hr/resume/internal/services/resume_service"
	"google.golang.org/grpc/codes"
)

func (a *ResumeServiceAPI) CreateCandidateFromResume(ctx context.Context, req *pb_models.CreateCandidateFromResumeRequest) (*pb_models.CandidateResumeResponse, error) {
	userCtx, err := getUserContext(ctx)
	if err != nil {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	res, err := a.resumeService.CreateCandidateFromResume(ctx, domain.CreateCandidateFromResumeInput{
		RequestUserID: userCtx.UserID,
		VacancyID:     req.GetVacancyId(),
		FileData:      req.GetFileData(),
	})
	if err != nil {
		switch {
		case errors.Is(err, resume_service.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid resume payload.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return &pb_models.CandidateResumeResponse{
		Candidate: toPBCandidate(*res.Candidate),
		Resume:    toPBResume(*res.Resume),
	}, nil
}
