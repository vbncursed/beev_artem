package auth_service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/auth/internal/domain"
)

type GetUserByIDSuite struct{ baseSuite }

func (s *GetUserByIDSuite) TestFound() {
	t := s.T()
	ctx := t.Context()
	want := &domain.User{ID: 99, Email: "x@example.com", Role: domain.RoleUser}

	s.authStorage.GetUserByIDMock.Expect(ctx, want.ID).Return(want, nil)

	got, err := s.svc.GetUserByID(ctx, want.ID)
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

func (s *GetUserByIDSuite) TestNotFound() {
	t := s.T()
	ctx := t.Context()

	s.authStorage.GetUserByIDMock.Expect(ctx, uint64(404)).Return(nil, nil)

	got, err := s.svc.GetUserByID(ctx, 404)
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.Assert(t, got == nil)
}

func (s *GetUserByIDSuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	storageErr := errors.New("pgx: connection refused")

	s.authStorage.GetUserByIDMock.Expect(ctx, uint64(1)).Return(nil, storageErr)

	got, err := s.svc.GetUserByID(ctx, 1)
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, got == nil)
}

func TestGetUserByIDSuite(t *testing.T) { suite.Run(t, new(GetUserByIDSuite)) }
