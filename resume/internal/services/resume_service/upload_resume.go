package resume_service

import (
	"context"
	"strings"

	"github.com/artem13815/hr/resume/internal/domain"
	"github.com/artem13815/hr/resume/internal/services/resume_service/extractor"
)

const maxResumeSizeBytes = 10 * 1024 * 1024

func (s *ResumeService) UploadResume(ctx context.Context, in domain.UploadResumeInput) (*domain.Resume, error) {
	if in.RequestUserID == 0 || strings.TrimSpace(in.CandidateID) == "" {
		return nil, ErrInvalidArgument
	}

	if len(in.Data) == 0 || len(in.Data) > maxResumeSizeBytes {
		return nil, ErrInvalidArgument
	}

	detectedType, err := extractor.DetectFileType(in.FileName, in.Data)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	in.FileType = detectedType
	in.FileName = extractor.BuildFileName(in.FileName, in.FileType)
	extractedText, err := extractor.ExtractText(in.FileType, in.Data)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	in.ExtractedText = extractedText

	resume, err := s.storage.UploadResume(ctx, in)
	if err != nil {
		return nil, err
	}
	if resume == nil {
		return nil, ErrNotFound
	}

	return resume, nil
}
