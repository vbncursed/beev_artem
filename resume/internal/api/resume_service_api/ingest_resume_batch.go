package resume_service_api

import (
	"context"
	"errors"

	"github.com/artem13815/hr/resume/internal/domain"
	pb_models "github.com/artem13815/hr/resume/internal/pb/models"
	"github.com/artem13815/hr/resume/internal/services/resume_service"
	"google.golang.org/grpc/codes"
)

func (a *ResumeServiceAPI) IngestResumeBatch(ctx context.Context, req *pb_models.BatchIngestResumeRequest) (*pb_models.BatchIngestResumeResponse, error) {
	userCtx, err := getUserContext(ctx)
	if err != nil {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	files := make([]domain.ResumeIntakeFile, 0, len(req.GetFiles()))
	for _, f := range req.GetFiles() {
		files = append(files, domain.ResumeIntakeFile{
			ExternalID: f.GetExternalId(),
			FileData:   f.GetFileData(),
		})
	}

	res, err := a.resumeService.IngestResumeBatch(ctx, domain.BatchIngestResumeInput{
		RequestUserID: userCtx.UserID,
		Files:         files,
	})
	if err != nil {
		switch {
		case errors.Is(err, resume_service.ErrInvalidArgument):
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid batch payload.")
		default:
			return nil, newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	out := make([]*pb_models.BatchIngestResumeItemResult, 0, len(res.Results))
	for _, item := range res.Results {
		pbItem := &pb_models.BatchIngestResumeItemResult{
			ExternalId: item.ExternalID,
			Error:      item.Error,
		}
		if item.Candidate != nil {
			pbItem.Candidate = toPBCandidate(*item.Candidate)
		}
		if item.Resume != nil {
			pbItem.Resume = toPBResume(*item.Resume)
		}
		out = append(out, pbItem)
	}

	return &pb_models.BatchIngestResumeResponse{Results: out}, nil
}
