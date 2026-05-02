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

func (a *ResumeServiceAPI) DownloadResume(ctx context.Context, req *pb_models.DownloadResumeRequest) (*pb_models.DownloadResumeResponse, error) {
	userCtx, ok := middleware.Get(ctx)
	if !ok {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	file, err := a.resumeService.DownloadResume(ctx, domain.DownloadResumeInput{
		RequestUserID: userCtx.UserID,
		IsAdmin:       userCtx.IsAdmin,
		ResumeID:      req.GetResumeId(),
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid resume id.")
		case errors.Is(err, usecase.ErrNotFound):
			return nil, newError(codes.NotFound, ErrCodeNotFound, "Resume not found.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return &pb_models.DownloadResumeResponse{
		FileData: file.Data,
		FileName: file.FileName,
		FileType: file.FileType,
	}, nil
}
