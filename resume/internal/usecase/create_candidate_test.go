package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/resume/internal/domain"
)

type CreateCandidateSuite struct{ baseSuite }

func (s *CreateCandidateSuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()

	in := domain.CreateCandidateInput{
		RequestUserID: 7,
		VacancyID:     "vac-1",
		FullName:      "Jane Doe",
		Email:         "jane@example.com",
	}
	want := &domain.Candidate{ID: "c-1", VacancyID: "vac-1", FullName: "Jane Doe"}

	s.storage.CreateCandidateMock.Expect(ctx, in).Return(want, nil)

	got, err := s.svc.CreateCandidate(ctx, in)
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

func (s *CreateCandidateSuite) TestInvalidArgumentZeroUser() {
	t := s.T()
	in := domain.CreateCandidateInput{VacancyID: "vac-1", FullName: "Jane"}

	got, err := s.svc.CreateCandidate(t.Context(), in)
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *CreateCandidateSuite) TestInvalidArgumentEmptyVacancy() {
	t := s.T()
	in := domain.CreateCandidateInput{RequestUserID: 1, VacancyID: "   ", FullName: "Jane"}

	got, err := s.svc.CreateCandidate(t.Context(), in)
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *CreateCandidateSuite) TestInvalidArgumentEmptyName() {
	t := s.T()
	in := domain.CreateCandidateInput{RequestUserID: 1, VacancyID: "vac-1", FullName: " "}

	got, err := s.svc.CreateCandidate(t.Context(), in)
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *CreateCandidateSuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	in := domain.CreateCandidateInput{RequestUserID: 1, VacancyID: "vac-1", FullName: "Jane"}
	storageErr := errors.New("pgx: deadlocked")

	s.storage.CreateCandidateMock.Expect(ctx, in).Return(nil, storageErr)

	got, err := s.svc.CreateCandidate(ctx, in)
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, got == nil)
}

func TestCreateCandidateSuite(t *testing.T) { suite.Run(t, new(CreateCandidateSuite)) }
