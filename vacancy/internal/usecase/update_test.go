package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

type UpdateVacancySuite struct{ baseSuite }

func (s *UpdateVacancySuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	in := domain.UpdateVacancyInput{
		VacancyID:   "v-1",
		OwnerUserID: 1,
		Title:       "Senior Backend",
		Description: "updated",
		Skills:      []domain.SkillWeight{{Name: "Go", Weight: 0.7}},
	}
	want := &domain.Vacancy{ID: "v-1", Title: in.Title}

	s.storage.UpdateVacancyMock.Expect(ctx, in).Return(want, nil)

	got, err := s.svc.UpdateVacancy(ctx, in)
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

// TestNormalizesZeroWeightedSkills: same contract as CreateVacancy — zero
// weights get redistributed to 1/N before the storage call.
func (s *UpdateVacancySuite) TestNormalizesZeroWeightedSkills() {
	t := s.T()
	ctx := t.Context()
	in := domain.UpdateVacancyInput{
		VacancyID:   "v-1",
		OwnerUserID: 1,
		Title:       "Lead",
		Skills: []domain.SkillWeight{
			{Name: "Go"},
			{Name: "SQL"},
		},
	}
	expected := domain.UpdateVacancyInput{
		VacancyID:   "v-1",
		OwnerUserID: 1,
		Title:       "Lead",
		Skills: []domain.SkillWeight{
			{Name: "Go", Weight: 0.5},
			{Name: "SQL", Weight: 0.5},
		},
	}
	want := &domain.Vacancy{ID: "v-1"}

	s.storage.UpdateVacancyMock.Expect(ctx, expected).Return(want, nil)

	got, err := s.svc.UpdateVacancy(ctx, in)
	assert.NilError(t, err)
	assert.Equal(t, got, want)
}

func (s *UpdateVacancySuite) TestInvalidArgumentZeroOwner() {
	t := s.T()
	got, err := s.svc.UpdateVacancy(t.Context(), domain.UpdateVacancyInput{
		VacancyID: "v-1",
		Title:     "ok",
		Skills:    []domain.SkillWeight{{Name: "Go", Weight: 0.5}},
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *UpdateVacancySuite) TestInvalidArgumentEmptyID() {
	t := s.T()
	got, err := s.svc.UpdateVacancy(t.Context(), domain.UpdateVacancyInput{
		OwnerUserID: 1,
		Title:       "ok",
		Skills:      []domain.SkillWeight{{Name: "Go", Weight: 0.5}},
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *UpdateVacancySuite) TestInvalidArgumentEmptyTitle() {
	t := s.T()
	got, err := s.svc.UpdateVacancy(t.Context(), domain.UpdateVacancyInput{
		VacancyID:   "v-1",
		OwnerUserID: 1,
		Title:       "   ",
		Skills:      []domain.SkillWeight{{Name: "Go", Weight: 0.5}},
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *UpdateVacancySuite) TestInvalidArgumentNoSkills() {
	t := s.T()
	got, err := s.svc.UpdateVacancy(t.Context(), domain.UpdateVacancyInput{
		VacancyID:   "v-1",
		OwnerUserID: 1,
		Title:       "ok",
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

// TestNotFoundOnNilStorageResult: storage returns (nil, nil) when the
// optimistic UPDATE matched zero rows (ownership mismatch or vacancy gone).
// Service translates this into ErrVacancyNotFound.
func (s *UpdateVacancySuite) TestNotFoundOnNilStorageResult() {
	t := s.T()
	ctx := t.Context()
	in := domain.UpdateVacancyInput{
		VacancyID:   "missing",
		OwnerUserID: 1,
		Title:       "ok",
		Skills:      []domain.SkillWeight{{Name: "Go", Weight: 0.5}},
	}

	s.storage.UpdateVacancyMock.Expect(ctx, in).Return(nil, nil)

	got, err := s.svc.UpdateVacancy(ctx, in)
	assert.ErrorIs(t, err, ErrVacancyNotFound)
	assert.Assert(t, got == nil)
}

func (s *UpdateVacancySuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	in := domain.UpdateVacancyInput{
		VacancyID:   "v-1",
		OwnerUserID: 1,
		Title:       "ok",
		Skills:      []domain.SkillWeight{{Name: "Go", Weight: 0.5}},
	}
	storageErr := errors.New("pgx: connection refused")

	s.storage.UpdateVacancyMock.Expect(ctx, in).Return(nil, storageErr)

	got, err := s.svc.UpdateVacancy(ctx, in)
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, got == nil)
}

func TestUpdateVacancySuite(t *testing.T) { suite.Run(t, new(UpdateVacancySuite)) }
