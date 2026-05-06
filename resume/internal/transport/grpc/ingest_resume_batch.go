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

func (a *ResumeServiceAPI) IngestResumeBatch(ctx context.Context, req *pb_models.BatchIngestResumeRequest) (*pb_models.BatchIngestResumeResponse, error) {
	userCtx, ok := middleware.Get(ctx)
	if !ok {
		return nil, newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	// Boundary checks before service-layer work: empty / oversized batch and
	// per-file size are all caller-provided and cheap to verify here. Service
	// still re-checks (defense in depth).
	if req.GetVacancyId() == "" {
		return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Vacancy ID is required.")
	}

	pbFiles := req.GetFiles()
	if len(pbFiles) == 0 || len(pbFiles) > usecase.MaxBatchFiles {
		return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "Batch size out of range.")
	}
	for _, f := range pbFiles {
		if size := len(f.GetFileData()); size == 0 || size > usecase.MaxResumeSizeBytes {
			return nil, newError(codes.InvalidArgument, ErrCodeInvalidInput, "One or more files have invalid size.")
		}
	}

	files := make([]domain.ResumeIntakeFile, 0, len(pbFiles))
	for _, f := range pbFiles {
		files = append(files, domain.ResumeIntakeFile{
			ExternalID: f.GetExternalId(),
			FileData:   f.GetFileData(),
		})
	}

	res, err := a.resumeService.IngestResumeBatch(ctx, domain.BatchIngestResumeInput{
		RequestUserID: userCtx.UserID,
		VacancyID:     req.GetVacancyId(),
		Files:         files,
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidArgument):
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
