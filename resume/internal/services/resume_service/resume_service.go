package resume_service

import (
	"context"

	"github.com/artem13815/hr/resume/internal/domain"
)

type ResumeStorage interface {
	CreateCandidate(ctx context.Context, in domain.CreateCandidateInput) (*domain.Candidate, error)
	GetCandidate(ctx context.Context, candidateID string, requestUserID uint64, isAdmin bool) (*domain.Candidate, error)
	UploadResume(ctx context.Context, in domain.UploadResumeInput) (*domain.Resume, error)
	GetResume(ctx context.Context, resumeID string, requestUserID uint64, isAdmin bool) (*domain.Resume, error)
}

type ResumeService struct {
	storage ResumeStorage
}

func NewResumeService(storage ResumeStorage) *ResumeService {
	return &ResumeService{storage: storage}
}
