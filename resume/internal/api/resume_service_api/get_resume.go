package resume_service_api

import (
	"context"
	"errors"

	"github.com/artem13815/hr/resume/internal/domain"
	pb_models "github.com/artem13815/hr/resume/internal/pb/models"
	"github.com/artem13815/hr/resume/internal/services/resume_service"
	"google.golang.org/grpc/codes"
)

func (a *ResumeServiceAPI) GetResume(ctx context.Context, req *pb_models.GetResumeRequest) (*pb_models.ResumeResponse, error) {
	userCtx, err := getUserContext(ctx)
	if err != nil {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	resume, err := a.resumeService.GetResume(ctx, domain.GetResumeInput{
		RequestUserID: userCtx.UserID,
		IsAdmin:       userCtx.IsAdmin,
		ResumeID:      req.GetResumeId(),
	})
	if err != nil {
		switch {
		case errors.Is(err, resume_service.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid resume id.")
		case errors.Is(err, resume_service.ErrNotFound):
			return nil, newError(codes.NotFound, ErrCodeNotFound, "Resume not found.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return &pb_models.ResumeResponse{Resume: toPBResume(*resume)}, nil
}
