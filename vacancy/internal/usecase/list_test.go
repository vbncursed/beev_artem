package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

type ListVacanciesSuite struct{ baseSuite }

func (s *ListVacanciesSuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	in := domain.ListVacanciesInput{
		OwnerUserID: 1,
		Limit:       50,
		Offset:      10,
		Query:       "go",
	}
	want := &domain.ListVacanciesResult{
		Vacancies: []domain.Vacancy{{ID: "v-1"}},
		Total:     1,
	}

	s.storage.ListVacanciesMock.Expect(ctx, in).Return(want, nil)

	got, err := s.svc.ListVacancies(ctx, in)
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

func (s *ListVacanciesSuite) TestUnauthorizedZeroOwner() {
	t := s.T()
	got, err := s.svc.ListVacancies(t.Context(), domain.ListVacanciesInput{Limit: 10})
	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Assert(t, got == nil)
}

// TestZeroLimitDefaultsTo20 covers cmp.Or behaviour: empty Limit becomes 20
// before the storage call.
func (s *ListVacanciesSuite) TestZeroLimitDefaultsTo20() {
	t := s.T()
	ctx := t.Context()
	want := &domain.ListVacanciesResult{Vacancies: []domain.Vacancy{}}
	expected := domain.ListVacanciesInput{OwnerUserID: 1, Limit: 20}

	s.storage.ListVacanciesMock.Expect(ctx, expected).Return(want, nil)

	got, err := s.svc.ListVacancies(ctx, domain.ListVacanciesInput{OwnerUserID: 1})
	assert.NilError(t, err)
	assert.Equal(t, got, want)
}

// TestLimitOver100Capped covers min(_, 100) behaviour: any caller-supplied
// page size larger than 100 is clamped before hitting storage. Protects
// against unbounded result sets.
func (s *ListVacanciesSuite) TestLimitOver100Capped() {
	t := s.T()
	ctx := t.Context()
	want := &domain.ListVacanciesResult{Vacancies: []domain.Vacancy{}}
	expected := domain.ListVacanciesInput{OwnerUserID: 1, Limit: 100}

	s.storage.ListVacanciesMock.Expect(ctx, expected).Return(want, nil)

	got, err := s.svc.ListVacancies(ctx, domain.ListVacanciesInput{OwnerUserID: 1, Limit: 9999})
	assert.NilError(t, err)
	assert.Equal(t, got, want)
}

func (s *ListVacanciesSuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	storageErr := errors.New("pgx: connection refused")
	expected := domain.ListVacanciesInput{OwnerUserID: 1, Limit: 20}

	s.storage.ListVacanciesMock.Expect(ctx, expected).Return(nil, storageErr)

	got, err := s.svc.ListVacancies(ctx, domain.ListVacanciesInput{OwnerUserID: 1})
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, got == nil)
}

func TestListVacanciesSuite(t *testing.T) { suite.Run(t, new(ListVacanciesSuite)) }
