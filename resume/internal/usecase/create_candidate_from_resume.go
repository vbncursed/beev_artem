package usecase

import (
	"context"
	"strings"

	"github.com/artem13815/hr/resume/internal/domain"
)

func (s *ResumeService) CreateCandidateFromResume(ctx context.Context, in domain.CreateCandidateFromResumeInput) (*domain.CandidateResumeResult, error) {
	if in.RequestUserID == 0 {
		return nil, ErrInvalidArgument
	}

	if len(in.FileData) == 0 || len(in.FileData) > MaxResumeSizeBytes {
		return nil, ErrInvalidArgument
	}

	fileType, err := s.extractor.DetectFileType("", in.FileData)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	fileName := s.extractor.BuildFileName("", fileType)

	extractedText, err := s.extractor.ExtractText(fileType, in.FileData)
	if err != nil {
		return nil, ErrInvalidArgument
	}

	profile := s.profile.ExtractCandidateProfile(extractedText, fileName)
	vacancyID := strings.TrimSpace(in.VacancyID)

	// Single transaction in storage: the candidate INSERT and resume INSERT
	// commit together or roll back together — no orphaned candidates if the
	// resume insert fails.
	candidate, resume, err := s.storage.CreateCandidateWithResume(ctx,
		domain.CreateCandidateInput{
			RequestUserID: in.RequestUserID,
			VacancyID:     vacancyID,
			FullName:      profile.FullName,
			Email:         profile.Email,
			Phone:         profile.Phone,
			Source:        domain.SourceFor(vacancyID),
		},
		domain.NewResumeData{
			FileName:      fileName,
			FileType:      fileType,
			ExtractedText: extractedText,
			Data:          in.FileData,
		},
	)
	if err != nil {
		return nil, err
	}

	return &domain.CandidateResumeResult{Candidate: candidate, Resume: resume}, nil
}
