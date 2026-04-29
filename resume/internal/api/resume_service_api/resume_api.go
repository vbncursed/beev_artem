package resume_service_api

import (
	"context"

	"github.com/artem13815/hr/resume/internal/domain"
	"github.com/artem13815/hr/resume/internal/pb/resume_api"
)

type resumeService interface {
	CreateCandidate(ctx context.Context, in domain.CreateCandidateInput) (*domain.Candidate, error)
	CreateCandidateFromResume(ctx context.Context, in domain.CreateCandidateFromResumeInput) (*domain.CandidateResumeResult, error)
	IngestResumeBatch(ctx context.Context, in domain.BatchIngestResumeInput) (*domain.BatchIngestResumeResult, error)
	GetCandidate(ctx context.Context, in domain.GetCandidateInput) (*domain.Candidate, error)
	UploadResume(ctx context.Context, in domain.UploadResumeInput) (*domain.Resume, error)
	GetResume(ctx context.Context, in domain.GetResumeInput) (*domain.Resume, error)
}

type ResumeServiceAPI struct {
	resume_api.UnimplementedResumeServiceServer
	resumeService resumeService
}

func NewResumeServiceAPI(service resumeService) *ResumeServiceAPI {
	return &ResumeServiceAPI{resumeService: service}
}

var _ resume_api.ResumeServiceServer = (*ResumeServiceAPI)(nil)
