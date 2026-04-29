package auth_service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/auth/internal/domain"
)

type LogoutSuite struct{ baseSuite }

func (s *LogoutSuite) TestEmptyToken() {
	t := s.T()
	err := s.svc.Logout(t.Context(), 1, "")
	assert.ErrorIs(t, err, ErrInvalidArgument)
}

func (s *LogoutSuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	tok := "ok-refresh"
	hash := hashRefreshToken(tok)

	s.sessionStorage.GetSessionByRefreshHashMock.Expect(ctx, hash).Return(&domain.Session{UserID: 42, ExpiresAt: time.Now().Add(time.Hour)}, nil)
	s.sessionStorage.RevokeSessionByRefreshHashMock.Expect(ctx, hash).Return(nil)

	assert.NilError(t, s.svc.Logout(ctx, 42, tok))
}

func (s *LogoutSuite) TestSessionNotFound() {
	t := s.T()
	ctx := t.Context()
	tok := "missing"

	s.sessionStorage.GetSessionByRefreshHashMock.Expect(ctx, hashRefreshToken(tok)).Return(nil, errors.New("not found"))

	err := s.svc.Logout(ctx, 1, tok)
	assert.ErrorIs(t, err, ErrInvalidRefreshToken)
}

func (s *LogoutSuite) TestSessionBelongsToOtherUser() {
	// H4 regression: when the supplied refresh token exists but belongs to a
	// different user, the service MUST return ErrInvalidRefreshToken (not
	// ErrPermissionDenied) so the response is indistinguishable from "not
	// found" — otherwise an attacker can probe whether a given token exists.
	t := s.T()
	ctx := t.Context()
	tok := "victims-refresh"
	hash := hashRefreshToken(tok)

	s.sessionStorage.GetSessionByRefreshHashMock.Expect(ctx, hash).Return(&domain.Session{UserID: 999}, nil)
	// RevokeSession must NOT be called — caller is not the owner.

	err := s.svc.Logout(ctx, 1, tok)
	assert.ErrorIs(t, err, ErrInvalidRefreshToken)
	assert.Assert(t, !errors.Is(err, ErrPermissionDenied))
}

func (s *LogoutSuite) TestRevokeStorageError() {
	t := s.T()
	ctx := t.Context()
	tok := "ok-refresh"
	hash := hashRefreshToken(tok)
	storageErr := errors.New("redis: timeout")

	s.sessionStorage.GetSessionByRefreshHashMock.Expect(ctx, hash).Return(&domain.Session{UserID: 7}, nil)
	s.sessionStorage.RevokeSessionByRefreshHashMock.Set(func(_ context.Context, _ []byte) error {
		return storageErr
	})

	err := s.svc.Logout(ctx, 7, tok)
	assert.ErrorIs(t, err, storageErr)
}

func TestLogoutSuite(t *testing.T) { suite.Run(t, new(LogoutSuite)) }
