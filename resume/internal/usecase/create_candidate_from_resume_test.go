package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/resume/internal/domain"
)

type CreateCandidateFromResumeSuite struct{ baseSuite }

// resumeText carries email/phone/name patterns that the multiagent extractor
// can pick up — the test asserts the extracted profile is propagated to
// storage, which exercises the extractor + multiagent pipeline as well.
var resumeText = []byte("Jane Doe\n" +
	"Email: jane@example.com\n" +
	"Phone +1 555 555 1234\n" +
	"Senior Engineer\n")

func (s *CreateCandidateFromResumeSuite) TestSuccessNoVacancy() {
	t := s.T()
	ctx := t.Context()

	wantCandidate := &domain.Candidate{ID: "c-1", FullName: "Jane Doe"}
	wantResume := &domain.Resume{ID: "r-1", CandidateID: "c-1"}

	s.storage.CreateCandidateWithResumeMock.Set(func(_ context.Context, candidateIn domain.CreateCandidateInput, resumeIn domain.NewResumeData) (*domain.Candidate, *domain.Resume, error) {
		assert.Equal(t, candidateIn.RequestUserID, uint64(7))
		assert.Equal(t, candidateIn.VacancyID, "")
		// Source defaults to resume_pool when no vacancy is supplied.
		assert.Equal(t, candidateIn.Source, "resume_pool")
		// Multiagent should have extracted email + name from resumeText.
		assert.Equal(t, candidateIn.Email, "jane@example.com")
		assert.Equal(t, candidateIn.FullName, "Jane Doe")
		assert.Equal(t, resumeIn.FileType, "txt")
		assert.Assert(t, resumeIn.ExtractedText != "")
		return wantCandidate, wantResume, nil
	})

	got, err := s.svc.CreateCandidateFromResume(ctx, domain.CreateCandidateFromResumeInput{
		RequestUserID: 7,
		FileData:      resumeText,
	})
	assert.NilError(t, err)
	assert.Equal(t, got.Candidate, wantCandidate)
	assert.Equal(t, got.Resume, wantResume)
}

func (s *CreateCandidateFromResumeSuite) TestSuccessWithVacancy() {
	t := s.T()
	ctx := t.Context()

	s.storage.CreateCandidateWithResumeMock.Set(func(_ context.Context, candidateIn domain.CreateCandidateInput, _ domain.NewResumeData) (*domain.Candidate, *domain.Resume, error) {
		assert.Equal(t, candidateIn.VacancyID, "vac-42")
		// Source flips to resume_auto when a vacancy is targeted.
		assert.Equal(t, candidateIn.Source, "resume_auto")
		return &domain.Candidate{ID: "c-2"}, &domain.Resume{ID: "r-2"}, nil
	})

	_, err := s.svc.CreateCandidateFromResume(ctx, domain.CreateCandidateFromResumeInput{
		RequestUserID: 7,
		VacancyID:     " vac-42 ",
		FileData:      resumeText,
	})
	assert.NilError(t, err)
}

func (s *CreateCandidateFromResumeSuite) TestInvalidArgumentZeroUser() {
	t := s.T()
	got, err := s.svc.CreateCandidateFromResume(t.Context(), domain.CreateCandidateFromResumeInput{FileData: resumeText})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *CreateCandidateFromResumeSuite) TestInvalidArgumentEmptyFile() {
	t := s.T()
	got, err := s.svc.CreateCandidateFromResume(t.Context(), domain.CreateCandidateFromResumeInput{
		RequestUserID: 1,
		FileData:      nil,
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *CreateCandidateFromResumeSuite) TestInvalidArgumentTooLarge() {
	t := s.T()
	tooBig := make([]byte, MaxResumeSizeBytes+1)
	got, err := s.svc.CreateCandidateFromResume(t.Context(), domain.CreateCandidateFromResumeInput{
		RequestUserID: 1,
		FileData:      tooBig,
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *CreateCandidateFromResumeSuite) TestInvalidArgumentUnknownFileType() {
	t := s.T()
	got, err := s.svc.CreateCandidateFromResume(t.Context(), domain.CreateCandidateFromResumeInput{
		RequestUserID: 1,
		FileData:      []byte{0x00, 0x01, 0x02, 0x03},
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *CreateCandidateFromResumeSuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	storageErr := errors.New("pgx: tx aborted")

	s.storage.CreateCandidateWithResumeMock.Set(func(_ context.Context, _ domain.CreateCandidateInput, _ domain.NewResumeData) (*domain.Candidate, *domain.Resume, error) {
		return nil, nil, storageErr
	})

	got, err := s.svc.CreateCandidateFromResume(ctx, domain.CreateCandidateFromResumeInput{
		RequestUserID: 1,
		FileData:      resumeText,
	})
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, got == nil)
}

func TestCreateCandidateFromResumeSuite(t *testing.T) {
	suite.Run(t, new(CreateCandidateFromResumeSuite))
}
