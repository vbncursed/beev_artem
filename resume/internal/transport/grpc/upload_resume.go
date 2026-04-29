package grpc

import (
	"bytes"
	"errors"
	"io"

	"github.com/artem13815/hr/resume/internal/domain"
	pb_models "github.com/artem13815/hr/resume/internal/pb/models"
	"github.com/artem13815/hr/resume/internal/pb/resume_api"
	"github.com/artem13815/hr/resume/internal/transport/middleware"
	"github.com/artem13815/hr/resume/internal/usecase"
	"google.golang.org/grpc/codes"
)

func (a *ResumeServiceAPI) UploadResume(stream resume_api.ResumeService_UploadResumeServer) error {
	userCtx, ok := middleware.Get(stream.Context())
	if !ok {
		return newError(codes.Unauthenticated, ErrCodeUnauthorized, "Authentication required.")
	}

	var meta *pb_models.UploadResumeMeta
	buf := bytes.NewBuffer(nil)
	buf.Grow(64 * 1024)

	for {
		req, recvErr := stream.Recv()
		if errors.Is(recvErr, io.EOF) {
			break
		}
		if recvErr != nil {
			return newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid stream payload.")
		}

		if m := req.GetMeta(); m != nil {
			if meta != nil {
				return newError(codes.InvalidArgument, ErrCodeInvalidInput, "Metadata already provided.")
			}
			meta = m
			continue
		}

		if c := req.GetChunk(); c != nil {
			if meta == nil {
				return newError(codes.InvalidArgument, ErrCodeInvalidInput, "Metadata must be sent first.")
			}
			data := c.GetData()
			if buf.Len()+len(data) > usecase.MaxResumeSizeBytes {
				return newError(codes.InvalidArgument, ErrCodeInvalidInput, "Resume exceeds maximum size.")
			}
			_, _ = buf.Write(data)
		}
	}

	if meta == nil {
		return newError(codes.InvalidArgument, ErrCodeInvalidInput, "Missing metadata.")
	}

	resume, err := a.resumeService.UploadResume(stream.Context(), domain.UploadResumeInput{
		RequestUserID: userCtx.UserID,
		IsAdmin:       userCtx.IsAdmin,
		CandidateID:   meta.GetCandidateId(),
		FileName:      meta.GetFileName(),
		FileType:      meta.GetFileType(),
		Data:          buf.Bytes(),
	})
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidArgument):
			return newError(codes.InvalidArgument, ErrCodeInvalidInput, "Invalid resume payload.")
		case errors.Is(err, usecase.ErrNotFound):
			return newError(codes.NotFound, ErrCodeNotFound, "Candidate not found.")
		default:
			return newError(codes.Internal, ErrCodeInternal, "Internal error.")
		}
	}

	return stream.SendAndClose(&pb_models.UploadResumeResponse{Resume: toPBResume(*resume)})
}
