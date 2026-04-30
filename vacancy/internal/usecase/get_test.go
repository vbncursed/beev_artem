package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

type GetVacancySuite struct{ baseSuite }

func (s *GetVacancySuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	want := &domain.Vacancy{ID: "v-1", OwnerUserID: 7}

	s.storage.GetVacancyMock.Expect(ctx, "v-1", uint64(7), false).Return(want, nil)

	got, err := s.svc.GetVacancy(ctx, domain.GetVacancyInput{
		VacancyID:   "v-1",
		OwnerUserID: 7,
	})
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

func (s *GetVacancySuite) TestSuccessAdmin() {
	t := s.T()
	ctx := t.Context()
	want := &domain.Vacancy{ID: "v-1", OwnerUserID: 99}

	s.storage.GetVacancyMock.Expect(ctx, "v-1", uint64(7), true).Return(want, nil)

	got, err := s.svc.GetVacancy(ctx, domain.GetVacancyInput{
		VacancyID:   "v-1",
		OwnerUserID: 7,
		IsAdmin:     true,
	})
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

func (s *GetVacancySuite) TestInvalidArgumentZeroOwner() {
	t := s.T()
	got, err := s.svc.GetVacancy(t.Context(), domain.GetVacancyInput{VacancyID: "v-1"})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *GetVacancySuite) TestInvalidArgumentEmptyID() {
	t := s.T()
	got, err := s.svc.GetVacancy(t.Context(), domain.GetVacancyInput{OwnerUserID: 1})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

// TestNotFoundOnNilStorageResult: storage returns (nil, nil) when ownership
// filter excludes the row — service must translate that to ErrVacancyNotFound
// instead of leaking nil to the handler layer.
func (s *GetVacancySuite) TestNotFoundOnNilStorageResult() {
	t := s.T()
	ctx := t.Context()

	s.storage.GetVacancyMock.Expect(ctx, "missing", uint64(1), false).Return(nil, nil)

	got, err := s.svc.GetVacancy(ctx, domain.GetVacancyInput{
		VacancyID:   "missing",
		OwnerUserID: 1,
	})
	assert.ErrorIs(t, err, ErrVacancyNotFound)
	assert.Assert(t, got == nil)
}

func (s *GetVacancySuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	storageErr := errors.New("pgx: connection refused")

	s.storage.GetVacancyMock.Expect(ctx, "v-1", uint64(1), false).Return(nil, storageErr)

	got, err := s.svc.GetVacancy(ctx, domain.GetVacancyInput{
		VacancyID:   "v-1",
		OwnerUserID: 1,
	})
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, got == nil)
}

func TestGetVacancySuite(t *testing.T) { suite.Run(t, new(GetVacancySuite)) }
