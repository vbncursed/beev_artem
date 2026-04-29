package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/auth/internal/domain"
)

type LoginSuite struct{ baseSuite }

func mustHash(t *testing.T, pw string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.MinCost) // MinCost for fast tests
	assert.NilError(t, err)
	return string(h)
}

func (s *LoginSuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	pw := "Password123!"
	user := &domain.User{ID: 7, Email: "ok@example.com", PasswordHash: mustHash(t, pw), Role: domain.RoleUser}
	in := domain.LoginInput{Email: user.Email, Password: pw, UserAgent: "ua", IP: "10.0.0.1"}

	s.authStorage.GetUserByEmailMock.Expect(ctx, in.Email).Return(user, nil)
	s.sessionStorage.CreateSessionMock.Inspect(func(_ context.Context, userID uint64, refreshHash []byte, _ time.Time, ua, ip string) {
		assert.Equal(t, userID, user.ID)
		assert.Equal(t, len(refreshHash), 32)
		assert.Equal(t, ua, in.UserAgent)
		assert.Equal(t, ip, in.IP)
	}).Return(nil)

	info, err := s.svc.Login(ctx, in)
	assert.NilError(t, err)
	assert.Equal(t, info.UserID, user.ID)
	assert.Assert(t, info.AccessToken != "")
	assert.Assert(t, info.RefreshToken != "")
}

func (s *LoginSuite) TestStorageErrorMaskedAsInvalidCredentials() {
	t := s.T()
	ctx := t.Context()
	in := domain.LoginInput{Email: "x@example.com", Password: "Password123!"}

	s.authStorage.GetUserByEmailMock.Expect(ctx, in.Email).Return(nil, errors.New("postgres down"))

	info, err := s.svc.Login(ctx, in)
	// Storage failures are intentionally collapsed to InvalidCredentials so
	// the response shape doesn't differ between "user missing" and "DB hiccup",
	// which would let an attacker probe enumeration.
	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Assert(t, info == nil)
}

func (s *LoginSuite) TestUserNotFound() {
	t := s.T()
	ctx := t.Context()
	in := domain.LoginInput{Email: "missing@example.com", Password: "Password123!"}

	s.authStorage.GetUserByEmailMock.Expect(ctx, in.Email).Return(nil, nil)

	info, err := s.svc.Login(ctx, in)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Assert(t, info == nil)
}

func (s *LoginSuite) TestWrongPassword() {
	t := s.T()
	ctx := t.Context()
	user := &domain.User{ID: 1, Email: "u@example.com", PasswordHash: mustHash(t, "RealPassword1!")}
	in := domain.LoginInput{Email: user.Email, Password: "WrongPassword1!"}

	s.authStorage.GetUserByEmailMock.Expect(ctx, in.Email).Return(user, nil)

	info, err := s.svc.Login(ctx, in)
	assert.ErrorIs(t, err, ErrInvalidCredentials)
	assert.Assert(t, info == nil)
}

func (s *LoginSuite) TestInvalidEmailShortCircuits() {
	t := s.T()
	ctx := t.Context()
	info, err := s.svc.Login(ctx, domain.LoginInput{Email: "bad", Password: "Password123!"})
	assert.ErrorIs(t, err, ErrInvalidEmail)
	assert.Assert(t, info == nil)
}

func TestLoginSuite(t *testing.T) { suite.Run(t, new(LoginSuite)) }
