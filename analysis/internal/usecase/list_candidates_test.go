package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/analysis/internal/domain"
)

type ListCandidatesByVacancySuite struct{ baseSuite }

func (s *ListCandidatesByVacancySuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	in := domain.ListCandidatesByVacancyInput{
		RequestUserID: 1,
		VacancyID:     "v-1",
		Limit:         50,
		Offset:        10,
	}
	want := &domain.ListCandidatesByVacancyResult{
		Candidates: []domain.CandidateWithAnalysis{{CandidateID: "c-1"}},
		Total:      1,
	}

	s.storage.ListCandidatesByVacancyMock.Expect(ctx, in).Return(want, nil)

	got, err := s.svc.ListCandidatesByVacancy(ctx, in)
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

// TestZeroLimitDefaultsTo20 covers the cmp.Or branch: empty Limit becomes 20
// before the storage call. The mock expectation captures the *normalised*
// input — the storage never sees a 0.
func (s *ListCandidatesByVacancySuite) TestZeroLimitDefaultsTo20() {
	t := s.T()
	ctx := t.Context()
	want := &domain.ListCandidatesByVacancyResult{Candidates: []domain.CandidateWithAnalysis{}}
	expected := domain.ListCandidatesByVacancyInput{RequestUserID: 1, VacancyID: "v-1", Limit: 20}

	s.storage.ListCandidatesByVacancyMock.Expect(ctx, expected).Return(want, nil)

	got, err := s.svc.ListCandidatesByVacancy(ctx, domain.ListCandidatesByVacancyInput{
		RequestUserID: 1,
		VacancyID:     "v-1",
	})
	assert.NilError(t, err)
	assert.Equal(t, got, want)
}

func (s *ListCandidatesByVacancySuite) TestInvalidArgumentZeroUser() {
	t := s.T()
	got, err := s.svc.ListCandidatesByVacancy(t.Context(), domain.ListCandidatesByVacancyInput{VacancyID: "v-1"})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *ListCandidatesByVacancySuite) TestInvalidArgumentEmptyVacancy() {
	t := s.T()
	got, err := s.svc.ListCandidatesByVacancy(t.Context(), domain.ListCandidatesByVacancyInput{
		RequestUserID: 1,
		VacancyID:     "   ",
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *ListCandidatesByVacancySuite) TestNotFoundOnNilStorageResult() {
	t := s.T()
	ctx := t.Context()
	expected := domain.ListCandidatesByVacancyInput{RequestUserID: 1, VacancyID: "v-1", Limit: 20}

	s.storage.ListCandidatesByVacancyMock.Expect(ctx, expected).Return(nil, nil)

	got, err := s.svc.ListCandidatesByVacancy(ctx, domain.ListCandidatesByVacancyInput{
		RequestUserID: 1,
		VacancyID:     "v-1",
	})
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Assert(t, got == nil)
}

func (s *ListCandidatesByVacancySuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	storageErr := errors.New("pgx: connection refused")
	expected := domain.ListCandidatesByVacancyInput{RequestUserID: 1, VacancyID: "v-1", Limit: 20}

	s.storage.ListCandidatesByVacancyMock.Expect(ctx, expected).Return(nil, storageErr)

	got, err := s.svc.ListCandidatesByVacancy(ctx, domain.ListCandidatesByVacancyInput{
		RequestUserID: 1,
		VacancyID:     "v-1",
	})
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, got == nil)
}

func TestListCandidatesByVacancySuite(t *testing.T) {
	suite.Run(t, new(ListCandidatesByVacancySuite))
}
