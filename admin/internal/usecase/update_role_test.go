package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/admin/internal/domain"
)

type UpdateRoleSuite struct{ baseSuite }

func (s *UpdateRoleSuite) TestPromote() {
	t := s.T()
	ctx := t.Context()

	s.authClient.UpdateUserRoleMock.Expect(ctx, uint64(7), domain.RoleAdmin).Return(nil)

	err := s.svc.UpdateRole(ctx, domain.UpdateRoleInput{
		CallerUserID: 1,
		IsAdmin:      true,
		TargetUserID: 7,
		NewRole:      domain.RoleAdmin,
	})
	assert.NilError(t, err)
}

func (s *UpdateRoleSuite) TestDemote() {
	t := s.T()
	ctx := t.Context()

	s.authClient.UpdateUserRoleMock.Expect(ctx, uint64(7), domain.RoleUser).Return(nil)

	err := s.svc.UpdateRole(ctx, domain.UpdateRoleInput{
		CallerUserID: 1,
		IsAdmin:      true,
		TargetUserID: 7,
		NewRole:      domain.RoleUser,
	})
	assert.NilError(t, err)
}

func (s *UpdateRoleSuite) TestRejectsZeroTargetID() {
	t := s.T()
	err := s.svc.UpdateRole(t.Context(), domain.UpdateRoleInput{
		IsAdmin: true,
		NewRole: domain.RoleAdmin,
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
}

func (s *UpdateRoleSuite) TestRejectsUnknownRole() {
	t := s.T()
	err := s.svc.UpdateRole(t.Context(), domain.UpdateRoleInput{
		IsAdmin:      true,
		TargetUserID: 7,
		NewRole:      "superuser",
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
}

func (s *UpdateRoleSuite) TestRejectsEmptyRole() {
	t := s.T()
	err := s.svc.UpdateRole(t.Context(), domain.UpdateRoleInput{
		IsAdmin:      true,
		TargetUserID: 7,
		NewRole:      "",
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
}

// TestRejectsNonAdminCaller — defense in depth. Auth interceptor already
// blocks non-admin tokens at the transport layer; usecase still refuses to
// proxy the call, so a future caller path that bypasses the interceptor (e.g.
// internal CLI, queue worker) cannot escalate.
func (s *UpdateRoleSuite) TestRejectsNonAdminCaller() {
	t := s.T()
	err := s.svc.UpdateRole(t.Context(), domain.UpdateRoleInput{
		IsAdmin:      false,
		TargetUserID: 7,
		NewRole:      domain.RoleAdmin,
	})
	assert.ErrorIs(t, err, ErrUnauthorized)
}

func (s *UpdateRoleSuite) TestAuthClientError() {
	t := s.T()
	ctx := t.Context()
	authErr := errors.New("auth.UpdateUserRole: rpc error")

	s.authClient.UpdateUserRoleMock.Expect(ctx, uint64(7), domain.RoleAdmin).Return(authErr)

	err := s.svc.UpdateRole(ctx, domain.UpdateRoleInput{
		IsAdmin:      true,
		TargetUserID: 7,
		NewRole:      domain.RoleAdmin,
	})
	assert.ErrorIs(t, err, authErr)
}

func TestUpdateRoleSuite(t *testing.T) { suite.Run(t, new(UpdateRoleSuite)) }
