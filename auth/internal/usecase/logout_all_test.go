package usecase

import (
	"github.com/artem13815/hr/auth/internal/infrastructure/jwt"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/auth/internal/domain"
)

type LogoutAllSuite struct{ baseSuite }

func (s *LogoutAllSuite) TestEmptyToken() {
	t := s.T()
	err := s.svc.LogoutAll(t.Context(), 1, "")
	assert.ErrorIs(t, err, ErrInvalidArgument)
}

func (s *LogoutAllSuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	tok := "ok-refresh"
	hash := jwt.HashRefresh(tok)

	s.sessionStorage.GetSessionByRefreshHashMock.Expect(ctx, hash).Return(&domain.Session{UserID: 42}, nil)
	s.sessionStorage.RevokeAllSessionsByUserIDMock.Expect(ctx, uint64(42)).Return(nil)

	assert.NilError(t, s.svc.LogoutAll(ctx, 42, tok))
}

func (s *LogoutAllSuite) TestSessionNotFound() {
	t := s.T()
	ctx := t.Context()
	tok := "missing"

	s.sessionStorage.GetSessionByRefreshHashMock.Expect(ctx, jwt.HashRefresh(tok)).Return(nil, errors.New("not found"))

	err := s.svc.LogoutAll(ctx, 1, tok)
	assert.ErrorIs(t, err, ErrInvalidRefreshToken)
}

func (s *LogoutAllSuite) TestSessionBelongsToOtherUser() {
	// H4 regression — same as Logout: uniform error, no permission-denied leak.
	t := s.T()
	ctx := t.Context()
	tok := "victims-refresh"

	s.sessionStorage.GetSessionByRefreshHashMock.Expect(ctx, jwt.HashRefresh(tok)).Return(&domain.Session{UserID: 999}, nil)

	err := s.svc.LogoutAll(ctx, 1, tok)
	assert.ErrorIs(t, err, ErrInvalidRefreshToken)
	assert.Assert(t, !errors.Is(err, ErrPermissionDenied))
}

func (s *LogoutAllSuite) TestRevokeAllStorageError() {
	t := s.T()
	ctx := t.Context()
	tok := "ok"
	storageErr := errors.New("redis down")

	s.sessionStorage.GetSessionByRefreshHashMock.Expect(ctx, jwt.HashRefresh(tok)).Return(&domain.Session{UserID: 7}, nil)
	s.sessionStorage.RevokeAllSessionsByUserIDMock.Set(func(_ context.Context, _ uint64) error {
		return storageErr
	})

	err := s.svc.LogoutAll(ctx, 7, tok)
	assert.ErrorIs(t, err, storageErr)
}

func TestLogoutAllSuite(t *testing.T) { suite.Run(t, new(LogoutAllSuite)) }
