package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/admin/internal/domain"
)

type ListUsersSuite struct{ baseSuite }

func (s *ListUsersSuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	want := []domain.AdminUserView{
		{
			ID:                 1,
			Email:              "alice@example.com",
			Role:               "admin",
			CreatedAt:          time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
			VacanciesOwned:     7,
			CandidatesUploaded: 19,
		},
		{
			ID:                 2,
			Email:              "bob@example.com",
			Role:               "user",
			CreatedAt:          time.Date(2026, 2, 14, 0, 0, 0, 0, time.UTC),
			VacanciesOwned:     0,
			CandidatesUploaded: 0,
		},
	}

	s.storage.ListUsersMock.Expect(ctx).Return(want, nil)

	got, err := s.svc.ListUsers(ctx)
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

func (s *ListUsersSuite) TestEmpty() {
	t := s.T()
	ctx := t.Context()

	s.storage.ListUsersMock.Expect(ctx).Return([]domain.AdminUserView{}, nil)

	got, err := s.svc.ListUsers(ctx)
	assert.NilError(t, err)
	assert.Equal(t, len(got), 0)
}

func (s *ListUsersSuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	storageErr := errors.New("pgx: query failed")

	s.storage.ListUsersMock.Expect(ctx).Return(nil, storageErr)

	got, err := s.svc.ListUsers(ctx)
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, got == nil)
}

func TestListUsersSuite(t *testing.T) { suite.Run(t, new(ListUsersSuite)) }
