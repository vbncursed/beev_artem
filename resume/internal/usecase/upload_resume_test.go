package usecase

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/resume/internal/domain"
)

type UploadResumeSuite struct{ baseSuite }

// txtPayload is a UTF-8 text file the extractor accepts as plain text — no
// PDF/DOCX magic, no zip header. Extractor returns strings.TrimSpace of it.
var txtPayload = []byte("Resume of Jane Doe\nEmail jane@example.com\n+1 555 555 1234\n")

func (s *UploadResumeSuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()

	want := &domain.Resume{ID: "r-1", CandidateID: "c-1", FileName: "resume.txt", FileType: "txt"}

	// Service runs extractor before storage, so we match by candidate id and
	// admin flag — exact bytes / FileType / ExtractedText are derived inside
	// the service. Using a custom Mock matcher keeps the test readable.
	s.storage.UploadResumeMock.Set(func(_ context.Context, in domain.UploadResumeInput) (*domain.Resume, error) {
		assert.Equal(t, in.RequestUserID, uint64(7))
		assert.Equal(t, in.CandidateID, "c-1")
		assert.Equal(t, in.FileType, "txt")
		assert.Equal(t, in.FileName, "resume.txt")
		assert.Assert(t, bytes.Equal(in.Data, txtPayload))
		assert.Assert(t, in.ExtractedText != "")
		return want, nil
	})

	got, err := s.svc.UploadResume(ctx, domain.UploadResumeInput{
		RequestUserID: 7,
		CandidateID:   "c-1",
		FileName:      "resume.txt",
		Data:          txtPayload,
	})
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

func (s *UploadResumeSuite) TestInvalidArgumentZeroUser() {
	t := s.T()
	got, err := s.svc.UploadResume(t.Context(), domain.UploadResumeInput{CandidateID: "c-1", Data: txtPayload})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *UploadResumeSuite) TestInvalidArgumentEmptyCandidate() {
	t := s.T()
	got, err := s.svc.UploadResume(t.Context(), domain.UploadResumeInput{
		RequestUserID: 1, CandidateID: " ", Data: txtPayload,
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *UploadResumeSuite) TestInvalidArgumentEmptyData() {
	t := s.T()
	got, err := s.svc.UploadResume(t.Context(), domain.UploadResumeInput{
		RequestUserID: 1, CandidateID: "c-1", Data: nil,
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *UploadResumeSuite) TestInvalidArgumentTooLarge() {
	t := s.T()
	tooBig := make([]byte, MaxResumeSizeBytes+1)
	got, err := s.svc.UploadResume(t.Context(), domain.UploadResumeInput{
		RequestUserID: 1, CandidateID: "c-1", Data: tooBig,
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *UploadResumeSuite) TestInvalidArgumentUnknownFileType() {
	t := s.T()
	// Binary garbage with no PDF/DOCX/TXT signature — extractor rejects it.
	garbage := []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05}

	got, err := s.svc.UploadResume(t.Context(), domain.UploadResumeInput{
		RequestUserID: 1,
		CandidateID:   "c-1",
		Data:          garbage,
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *UploadResumeSuite) TestNotFoundFromStorage() {
	t := s.T()
	ctx := t.Context()

	s.storage.UploadResumeMock.Set(func(_ context.Context, _ domain.UploadResumeInput) (*domain.Resume, error) {
		return nil, domain.ErrNotFound
	})

	got, err := s.svc.UploadResume(ctx, domain.UploadResumeInput{
		RequestUserID: 1,
		CandidateID:   "missing",
		FileName:      "x.txt",
		Data:          txtPayload,
	})
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Assert(t, got == nil)
}

func (s *UploadResumeSuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	storageErr := errors.New("pgx: connection lost")

	s.storage.UploadResumeMock.Set(func(_ context.Context, _ domain.UploadResumeInput) (*domain.Resume, error) {
		return nil, storageErr
	})

	got, err := s.svc.UploadResume(ctx, domain.UploadResumeInput{
		RequestUserID: 1,
		CandidateID:   "c-1",
		FileName:      "r.txt",
		Data:          txtPayload,
	})
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, got == nil)
}

func TestUploadResumeSuite(t *testing.T) { suite.Run(t, new(UploadResumeSuite)) }
