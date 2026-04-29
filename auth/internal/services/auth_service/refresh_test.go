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

type RefreshSuite struct{ baseSuite }

func (s *RefreshSuite) TestEmptyToken() {
	t := s.T()
	ctx := t.Context()

	info, err := s.svc.Refresh(ctx, domain.RefreshInput{RefreshToken: ""})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, info == nil)
}

func (s *RefreshSuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	user := &domain.User{ID: 9, Email: "u@example.com", Role: domain.RoleUser}
	oldRefresh := "old-refresh-token-value"
	expectedHash := hashRefreshToken(oldRefresh)
	sess := &domain.Session{UserID: user.ID, ExpiresAt: time.Now().Add(time.Hour)}

	s.sessionStorage.ConsumeSessionByRefreshHashMock.Expect(ctx, expectedHash).Return(sess, nil)
	s.authStorage.GetUserByIDMock.Expect(ctx, user.ID).Return(user, nil)
	s.sessionStorage.CreateSessionMock.Inspect(func(_ context.Context, userID uint64, refreshHash []byte, _ time.Time, _, _ string) {
		assert.Equal(t, userID, user.ID)
		// New refresh hash MUST differ from the consumed one.
		assert.Assert(t, string(refreshHash) != string(expectedHash))
	}).Return(nil)

	info, err := s.svc.Refresh(ctx, domain.RefreshInput{RefreshToken: oldRefresh, UserAgent: "ua", IP: "1.1.1.1"})
	assert.NilError(t, err)
	assert.Equal(t, info.UserID, user.ID)
	assert.Assert(t, info.RefreshToken != oldRefresh)
}

func (s *RefreshSuite) TestSessionNotFoundOrAlreadyConsumed() {
	t := s.T()
	ctx := t.Context()
	tok := "some-refresh"

	s.sessionStorage.ConsumeSessionByRefreshHashMock.Expect(ctx, hashRefreshToken(tok)).Return(nil, errors.New("session: not found"))

	info, err := s.svc.Refresh(ctx, domain.RefreshInput{RefreshToken: tok})
	assert.ErrorIs(t, err, ErrInvalidRefreshToken)
	assert.Assert(t, info == nil)
}

func (s *RefreshSuite) TestSessionExpired() {
	t := s.T()
	ctx := t.Context()
	tok := "expired-refresh"
	sess := &domain.Session{UserID: 1, ExpiresAt: time.Now().Add(-time.Minute)}

	s.sessionStorage.ConsumeSessionByRefreshHashMock.Expect(ctx, hashRefreshToken(tok)).Return(sess, nil)

	info, err := s.svc.Refresh(ctx, domain.RefreshInput{RefreshToken: tok})
	assert.ErrorIs(t, err, ErrSessionExpired)
	assert.Assert(t, info == nil)
}

func (s *RefreshSuite) TestUserVanished() {
	t := s.T()
	ctx := t.Context()
	tok := "valid-refresh"
	sess := &domain.Session{UserID: 555, ExpiresAt: time.Now().Add(time.Hour)}

	s.sessionStorage.ConsumeSessionByRefreshHashMock.Expect(ctx, hashRefreshToken(tok)).Return(sess, nil)
	s.authStorage.GetUserByIDMock.Expect(ctx, sess.UserID).Return(nil, nil)

	info, err := s.svc.Refresh(ctx, domain.RefreshInput{RefreshToken: tok})
	assert.ErrorIs(t, err, ErrUserNotFound)
	assert.Assert(t, info == nil)
}

func (s *RefreshSuite) TestReplayBlocked() {
	// Regression test for C2: two sequential Refresh calls with the same
	// refresh token must NOT both succeed. The mock simulates GETDEL semantics
	// by returning the session on the first call and ErrSessionNotFound on
	// every subsequent call.
	t := s.T()
	ctx := t.Context()
	user := &domain.User{ID: 11, Email: "u@example.com", Role: domain.RoleUser}
	tok := "replayed-refresh"
	hash := hashRefreshToken(tok)

	consumed := false
	s.sessionStorage.ConsumeSessionByRefreshHashMock.Set(func(_ context.Context, refreshHash []byte) (*domain.Session, error) {
		assert.Equal(t, string(refreshHash), string(hash))
		if consumed {
			return nil, errors.New("session: not found")
		}
		consumed = true
		return &domain.Session{UserID: user.ID, ExpiresAt: time.Now().Add(time.Hour)}, nil
	})
	s.authStorage.GetUserByIDMock.Expect(ctx, user.ID).Return(user, nil)
	s.sessionStorage.CreateSessionMock.Set(func(_ context.Context, _ uint64, _ []byte, _ time.Time, _, _ string) error {
		return nil
	})

	info, err := s.svc.Refresh(ctx, domain.RefreshInput{RefreshToken: tok})
	assert.NilError(t, err)
	assert.Equal(t, info.UserID, user.ID)

	// Second attempt with the same token — replay must fail.
	info2, err := s.svc.Refresh(ctx, domain.RefreshInput{RefreshToken: tok})
	assert.ErrorIs(t, err, ErrInvalidRefreshToken)
	assert.Assert(t, info2 == nil)
}

func TestRefreshSuite(t *testing.T) { suite.Run(t, new(RefreshSuite)) }
