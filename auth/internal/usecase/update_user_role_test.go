package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/auth/internal/domain"
)

type UpdateUserRoleSuite struct{ baseSuite }

func (s *UpdateUserRoleSuite) TestAdminPromotesUser() {
	t := s.T()
	ctx := t.Context()
	admin := &domain.User{ID: 1, Role: domain.RoleAdmin}
	target := &domain.User{ID: 2, Role: domain.RoleUser}

	s.authStorage.GetUserByIDMock.When(ctx, admin.ID).Then(admin, nil)
	s.authStorage.GetUserByIDMock.When(ctx, target.ID).Then(target, nil)
	s.authStorage.UpdateUserRoleMock.Expect(ctx, target.ID, domain.RoleAdmin).Return(nil)

	assert.NilError(t, s.svc.UpdateUserRole(ctx, admin.ID, target.ID, domain.RoleAdmin))
}

func (s *UpdateUserRoleSuite) TestNonAdminDenied() {
	t := s.T()
	ctx := t.Context()
	caller := &domain.User{ID: 5, Role: domain.RoleUser}

	s.authStorage.GetUserByIDMock.Expect(ctx, caller.ID).Return(caller, nil)

	err := s.svc.UpdateUserRole(ctx, caller.ID, 99, domain.RoleAdmin)
	assert.ErrorIs(t, err, ErrPermissionDenied)
}

func (s *UpdateUserRoleSuite) TestSelfUpdateDenied() {
	t := s.T()
	ctx := t.Context()
	admin := &domain.User{ID: 3, Role: domain.RoleAdmin}

	s.authStorage.GetUserByIDMock.Expect(ctx, admin.ID).Return(admin, nil)

	err := s.svc.UpdateUserRole(ctx, admin.ID, admin.ID, domain.RoleUser)
	assert.ErrorIs(t, err, ErrCannotChangeOwnRole)
}

func (s *UpdateUserRoleSuite) TestInvalidRole() {
	t := s.T()
	ctx := t.Context()
	// Validation runs first — no storage call expected.
	err := s.svc.UpdateUserRole(ctx, 1, 2, "superuser")
	assert.ErrorIs(t, err, ErrInvalidRole)
}

func (s *UpdateUserRoleSuite) TestAdminNotFound() {
	t := s.T()
	ctx := t.Context()

	s.authStorage.GetUserByIDMock.Expect(ctx, uint64(7)).Return(nil, nil)

	err := s.svc.UpdateUserRole(ctx, 7, 8, domain.RoleAdmin)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func (s *UpdateUserRoleSuite) TestTargetNotFound() {
	t := s.T()
	ctx := t.Context()
	admin := &domain.User{ID: 1, Role: domain.RoleAdmin}

	s.authStorage.GetUserByIDMock.When(ctx, admin.ID).Then(admin, nil)
	s.authStorage.GetUserByIDMock.When(ctx, uint64(404)).Then(nil, nil)

	err := s.svc.UpdateUserRole(ctx, admin.ID, 404, domain.RoleUser)
	assert.ErrorIs(t, err, ErrUserNotFound)
}

func (s *UpdateUserRoleSuite) TestAdminLookupStorageError() {
	t := s.T()
	ctx := t.Context()
	storageErr := errors.New("postgres timeout")

	s.authStorage.GetUserByIDMock.Expect(ctx, uint64(1)).Return(nil, storageErr)

	err := s.svc.UpdateUserRole(ctx, 1, 2, domain.RoleUser)
	assert.ErrorIs(t, err, storageErr)
}

func TestUpdateUserRoleSuite(t *testing.T) { suite.Run(t, new(UpdateUserRoleSuite)) }
