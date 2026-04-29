package resume_service

import (
	"context"
	"strings"

	"github.com/artem13815/hr/resume/internal/domain"
	"github.com/artem13815/hr/resume/internal/services/resume_service/extractor"
	"github.com/artem13815/hr/resume/internal/services/resume_service/multiagent"
)

func (s *ResumeService) CreateCandidateFromResume(ctx context.Context, in domain.CreateCandidateFromResumeInput) (*domain.CandidateResumeResult, error) {
	if in.RequestUserID == 0 {
		return nil, ErrInvalidArgument
	}

	if len(in.FileData) == 0 || len(in.FileData) > maxResumeSizeBytes {
		return nil, ErrInvalidArgument
	}

	fileType, err := extractor.DetectFileType("", in.FileData)
	if err != nil {
		return nil, ErrInvalidArgument
	}
	fileName := extractor.BuildFileName("", fileType)

	extractedText, err := extractor.ExtractText(fileType, in.FileData)
	if err != nil {
		return nil, ErrInvalidArgument
	}

	profile := multiagent.ExtractCandidateProfile(extractedText, fileName)

	candidate, err := s.storage.CreateCandidate(ctx, domain.CreateCandidateInput{
		RequestUserID: in.RequestUserID,
		VacancyID:     strings.TrimSpace(in.VacancyID),
		FullName:      profile.FullName,
		Email:         profile.Email,
		Phone:         profile.Phone,
		Source:        candidateSource(strings.TrimSpace(in.VacancyID)),
		Comment:       "",
	})
	if err != nil {
		return nil, err
	}
	if candidate == nil {
		return nil, ErrNotFound
	}

	resume, err := s.storage.UploadResume(ctx, domain.UploadResumeInput{
		RequestUserID: in.RequestUserID,
		IsAdmin:       false,
		CandidateID:   candidate.ID,
		FileName:      fileName,
		FileType:      fileType,
		ExtractedText: extractedText,
		Data:          in.FileData,
	})
	if err != nil {
		return nil, err
	}
	if resume == nil {
		return nil, ErrNotFound
	}

	return &domain.CandidateResumeResult{Candidate: candidate, Resume: resume}, nil
}

func candidateSource(vacancyID string) string {
	if vacancyID == "" {
		return "resume_pool"
	}
	return "resume_auto"
}
