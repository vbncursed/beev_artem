package usecase

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

type CreateVacancySuite struct{ baseSuite }

// validSkills is a single, well-formed skill set that satisfies
// validateCreateInput (non-empty name, weight in [0,1]). Tests that need to
// exercise *other* validation rules use this so they do not fail on the
// skills check by accident.
func validSkills() []domain.SkillWeight {
	return []domain.SkillWeight{{Name: "Go", Weight: 0.5, MustHave: true}}
}

func (s *CreateVacancySuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	in := domain.CreateVacancyInput{
		OwnerUserID: 1,
		Title:       "Backend Engineer",
		Description: "We build things.",
		Skills:      validSkills(),
	}
	expected := in
	expected.Role = "programmer"
	want := &domain.Vacancy{ID: "v-1", OwnerUserID: 1, Title: in.Title}

	s.storage.CreateVacancyMock.Expect(ctx, expected).Return(want, nil)

	got, err := s.svc.CreateVacancy(ctx, in)
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

// TestNormalizesZeroWeightedSkills covers the contract that when every skill
// arrives with Weight=0, the service redistributes weights equally to 1/N
// before persisting. The mock expectation captures that the storage sees the
// *normalized* skills, never the raw zeros.
func (s *CreateVacancySuite) TestNormalizesZeroWeightedSkills() {
	t := s.T()
	ctx := t.Context()

	in := domain.CreateVacancyInput{
		OwnerUserID: 1,
		Title:       "Lead",
		Skills: []domain.SkillWeight{
			{Name: "Go"},
			{Name: "SQL"},
			{Name: "AWS"},
		},
	}
	expected := domain.CreateVacancyInput{
		OwnerUserID: 1,
		Title:       "Lead",
		Role:        "default",
		Skills: []domain.SkillWeight{
			{Name: "Go", Weight: float32(1) / 3},
			{Name: "SQL", Weight: float32(1) / 3},
			{Name: "AWS", Weight: float32(1) / 3},
		},
	}
	want := &domain.Vacancy{ID: "v-2"}

	s.storage.CreateVacancyMock.Expect(ctx, expected).Return(want, nil)

	got, err := s.svc.CreateVacancy(ctx, in)
	assert.NilError(t, err)
	assert.Equal(t, got, want)
}

func (s *CreateVacancySuite) TestUnauthorizedZeroOwner() {
	t := s.T()
	got, err := s.svc.CreateVacancy(t.Context(), domain.CreateVacancyInput{
		Title:  "x",
		Skills: validSkills(),
	})
	assert.ErrorIs(t, err, ErrUnauthorized)
	assert.Assert(t, got == nil)
}

func (s *CreateVacancySuite) TestInvalidArgumentEmptyTitle() {
	t := s.T()
	got, err := s.svc.CreateVacancy(t.Context(), domain.CreateVacancyInput{
		OwnerUserID: 1,
		Title:       "   ",
		Skills:      validSkills(),
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *CreateVacancySuite) TestInvalidArgumentTitleTooLong() {
	t := s.T()
	got, err := s.svc.CreateVacancy(t.Context(), domain.CreateVacancyInput{
		OwnerUserID: 1,
		Title:       strings.Repeat("a", 256),
		Skills:      validSkills(),
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *CreateVacancySuite) TestInvalidArgumentDescriptionTooLong() {
	t := s.T()
	got, err := s.svc.CreateVacancy(t.Context(), domain.CreateVacancyInput{
		OwnerUserID: 1,
		Title:       "ok",
		Description: strings.Repeat("a", 4001),
		Skills:      validSkills(),
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *CreateVacancySuite) TestInvalidArgumentNoSkills() {
	t := s.T()
	got, err := s.svc.CreateVacancy(t.Context(), domain.CreateVacancyInput{
		OwnerUserID: 1,
		Title:       "ok",
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *CreateVacancySuite) TestInvalidArgumentEmptySkillName() {
	t := s.T()
	got, err := s.svc.CreateVacancy(t.Context(), domain.CreateVacancyInput{
		OwnerUserID: 1,
		Title:       "ok",
		Skills:      []domain.SkillWeight{{Name: " ", Weight: 0.5}},
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *CreateVacancySuite) TestInvalidArgumentWeightOutOfRange() {
	t := s.T()
	got, err := s.svc.CreateVacancy(t.Context(), domain.CreateVacancyInput{
		OwnerUserID: 1,
		Title:       "ok",
		Skills:      []domain.SkillWeight{{Name: "Go", Weight: 1.5}},
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *CreateVacancySuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	in := domain.CreateVacancyInput{
		OwnerUserID: 1,
		Title:       "ok",
		Skills:      validSkills(),
	}
	expected := in
	expected.Role = "default"
	storageErr := errors.New("pgx: connection refused")

	s.storage.CreateVacancyMock.Expect(ctx, expected).Return(nil, storageErr)

	got, err := s.svc.CreateVacancy(ctx, in)
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, got == nil)
}

func TestCreateVacancySuite(t *testing.T) { suite.Run(t, new(CreateVacancySuite)) }
