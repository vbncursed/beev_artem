package usecase

//go:generate go run github.com/gojuno/minimock/v3/cmd/minimock@v3.4.7 -i ResumeStorage,TextExtractor,ProfileExtractor -o ./mocks -s _mock.go -g

import (
	"context"

	"github.com/artem13815/hr/resume/internal/domain"
)

// ResumeStorage is the persistence driven port. Implemented by
// infrastructure/persistence (pgx + goose).
type ResumeStorage interface {
	CreateCandidate(ctx context.Context, in domain.CreateCandidateInput) (*domain.Candidate, error)
	GetCandidate(ctx context.Context, candidateID string, requestUserID uint64, isAdmin bool) (*domain.Candidate, error)
	UploadResume(ctx context.Context, in domain.UploadResumeInput) (*domain.Resume, error)
	GetResume(ctx context.Context, resumeID string, requestUserID uint64, isAdmin bool) (*domain.Resume, error)

	// CreateCandidateWithResume inserts a candidate and its first resume in a
	// single transaction. Either both rows land or neither does — protects
	// against orphaned candidates when the resume insert fails after the
	// candidate insert succeeded.
	CreateCandidateWithResume(ctx context.Context, candidateIn domain.CreateCandidateInput, resumeIn domain.NewResumeData) (*domain.Candidate, *domain.Resume, error)
}

// TextExtractor is the file-format detection + text extraction driven port.
// Implemented by infrastructure/extractor (PDF / DOCX / TXT).
type TextExtractor interface {
	DetectFileType(fileName string, data []byte) (string, error)
	ExtractText(fileType string, data []byte) (string, error)
	BuildFileName(fileName, fileType string) string
}

// ProfileExtractor is the structured-profile extraction driven port.
// Implemented by infrastructure/profile (regex-based today, swappable for
// an ML/LLM extractor tomorrow without touching the use case).
type ProfileExtractor interface {
	ExtractCandidateProfile(text, fileName string) domain.CandidateProfile
}

type ResumeService struct {
	storage   ResumeStorage
	extractor TextExtractor
	profile   ProfileExtractor
}

func NewResumeService(storage ResumeStorage, extractor TextExtractor, profile ProfileExtractor) *ResumeService {
	return &ResumeService{
		storage:   storage,
		extractor: extractor,
		profile:   profile,
	}
}
