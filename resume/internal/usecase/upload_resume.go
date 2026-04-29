package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/artem13815/hr/resume/internal/domain"
)

// MaxResumeSizeBytes caps the size of an uploaded resume file.
// Keep in sync with grpc.MaxRecvMsgSize in bootstrap/server.go.
const MaxResumeSizeBytes = 10 * 1024 * 1024

func (s *ResumeService) UploadResume(ctx context.Context, in domain.UploadResumeInput) (*domain.Resume, error) {
	if in.RequestUserID == 0 || strings.TrimSpace(in.CandidateID) == "" {
		return nil, ErrInvalidArgument
	}

	if len(in.Data) == 0 || len(in.Data) > MaxResumeSizeBytes {
		return nil, ErrInvalidArgument
	}

	detectedType, err := s.extractor.DetectFileType(in.FileName, in.Data)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	in.FileType = detectedType
	in.FileName = s.extractor.BuildFileName(in.FileName, in.FileType)
	extractedText, err := s.extractor.ExtractText(in.FileType, in.Data)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	in.ExtractedText = extractedText

	resume, err := s.storage.UploadResume(ctx, in)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return resume, nil
}
