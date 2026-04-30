package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/vacancy/internal/domain"
)

type ArchiveVacancySuite struct{ baseSuite }

func (s *ArchiveVacancySuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	in := domain.ArchiveVacancyInput{VacancyID: "v-1", OwnerUserID: 7}

	s.storage.ArchiveVacancyMock.Expect(ctx, in).Return(nil)

	err := s.svc.ArchiveVacancy(ctx, in)
	assert.NilError(t, err)
}

func (s *ArchiveVacancySuite) TestSuccessAdmin() {
	t := s.T()
	ctx := t.Context()
	in := domain.ArchiveVacancyInput{VacancyID: "v-1", OwnerUserID: 7, IsAdmin: true}

	s.storage.ArchiveVacancyMock.Expect(ctx, in).Return(nil)

	err := s.svc.ArchiveVacancy(ctx, in)
	assert.NilError(t, err)
}

func (s *ArchiveVacancySuite) TestInvalidArgumentZeroOwner() {
	t := s.T()
	err := s.svc.ArchiveVacancy(t.Context(), domain.ArchiveVacancyInput{VacancyID: "v-1"})
	assert.ErrorIs(t, err, ErrInvalidArgument)
}

func (s *ArchiveVacancySuite) TestInvalidArgumentEmptyID() {
	t := s.T()
	err := s.svc.ArchiveVacancy(t.Context(), domain.ArchiveVacancyInput{OwnerUserID: 1})
	assert.ErrorIs(t, err, ErrInvalidArgument)
}

func (s *ArchiveVacancySuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	in := domain.ArchiveVacancyInput{VacancyID: "v-1", OwnerUserID: 1}
	storageErr := errors.New("pgx: connection refused")

	s.storage.ArchiveVacancyMock.Expect(ctx, in).Return(storageErr)

	err := s.svc.ArchiveVacancy(ctx, in)
	assert.ErrorIs(t, err, storageErr)
}

func TestArchiveVacancySuite(t *testing.T) { suite.Run(t, new(ArchiveVacancySuite)) }
