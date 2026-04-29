package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/resume/internal/domain"
)

type GetResumeSuite struct{ baseSuite }

func (s *GetResumeSuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	want := &domain.Resume{ID: "r-1", CandidateID: "c-1"}

	s.storage.GetResumeMock.Expect(ctx, "r-1", uint64(7), false).Return(want, nil)

	got, err := s.svc.GetResume(ctx, domain.GetResumeInput{RequestUserID: 7, ResumeID: "r-1"})
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

func (s *GetResumeSuite) TestInvalidArgument() {
	t := s.T()
	got, err := s.svc.GetResume(t.Context(), domain.GetResumeInput{ResumeID: "r-1"})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *GetResumeSuite) TestNotFound() {
	t := s.T()
	ctx := t.Context()

	s.storage.GetResumeMock.Expect(ctx, "missing", uint64(1), false).Return(nil, domain.ErrNotFound)

	got, err := s.svc.GetResume(ctx, domain.GetResumeInput{RequestUserID: 1, ResumeID: "missing"})
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Assert(t, got == nil)
}

func (s *GetResumeSuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	storageErr := errors.New("pgx: bad query")

	s.storage.GetResumeMock.Expect(ctx, "r-1", uint64(1), false).Return(nil, storageErr)

	got, err := s.svc.GetResume(ctx, domain.GetResumeInput{RequestUserID: 1, ResumeID: "r-1"})
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, got == nil)
}

func TestGetResumeSuite(t *testing.T) { suite.Run(t, new(GetResumeSuite)) }
