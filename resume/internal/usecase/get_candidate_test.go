package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/resume/internal/domain"
)

type GetCandidateSuite struct{ baseSuite }

func (s *GetCandidateSuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	want := &domain.Candidate{ID: "c-1", OwnerUserID: 7}

	s.storage.GetCandidateMock.Expect(ctx, "c-1", uint64(7), false).Return(want, nil)

	got, err := s.svc.GetCandidate(ctx, domain.GetCandidateInput{
		RequestUserID: 7,
		CandidateID:   "c-1",
	})
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

func (s *GetCandidateSuite) TestSuccessAdmin() {
	t := s.T()
	ctx := t.Context()
	want := &domain.Candidate{ID: "c-1", OwnerUserID: 99}

	s.storage.GetCandidateMock.Expect(ctx, "c-1", uint64(7), true).Return(want, nil)

	got, err := s.svc.GetCandidate(ctx, domain.GetCandidateInput{
		RequestUserID: 7,
		IsAdmin:       true,
		CandidateID:   "c-1",
	})
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

func (s *GetCandidateSuite) TestInvalidArgumentZeroUser() {
	t := s.T()
	got, err := s.svc.GetCandidate(t.Context(), domain.GetCandidateInput{CandidateID: "c-1"})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *GetCandidateSuite) TestInvalidArgumentEmptyID() {
	t := s.T()
	got, err := s.svc.GetCandidate(t.Context(), domain.GetCandidateInput{RequestUserID: 1, CandidateID: " "})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *GetCandidateSuite) TestNotFound() {
	t := s.T()
	ctx := t.Context()

	s.storage.GetCandidateMock.Expect(ctx, "missing", uint64(1), false).Return(nil, domain.ErrNotFound)

	got, err := s.svc.GetCandidate(ctx, domain.GetCandidateInput{
		RequestUserID: 1,
		CandidateID:   "missing",
	})
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Assert(t, got == nil)
}

func (s *GetCandidateSuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	storageErr := errors.New("pgx: connection refused")

	s.storage.GetCandidateMock.Expect(ctx, "c-1", uint64(1), false).Return(nil, storageErr)

	got, err := s.svc.GetCandidate(ctx, domain.GetCandidateInput{
		RequestUserID: 1,
		CandidateID:   "c-1",
	})
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, got == nil)
}

func TestGetCandidateSuite(t *testing.T) { suite.Run(t, new(GetCandidateSuite)) }
