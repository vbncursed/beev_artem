package auth_service

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

type RegisterSuite struct{ baseSuite }

func (s *RegisterSuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	in := domain.RegisterInput{Email: "new@example.com", Password: "Password123!", UserAgent: "ua", IP: "1.2.3.4"}

	s.authStorage.GetUserByEmailMock.Expect(ctx, in.Email).Return(nil, nil)

	// CreateUser receives the bcrypt hash of in.Password — we can't predict its
	// bytes (salt is random), so use Set to inspect the hash dynamically.
	s.authStorage.CreateUserMock.Set(func(_ context.Context, email, hash string) (uint64, error) {
		assert.Equal(t, email, in.Email)
		assert.NilError(t, bcrypt.CompareHashAndPassword([]byte(hash), []byte(in.Password)))
		return 42, nil
	})

	s.sessionStorage.CreateSessionMock.Inspect(func(_ context.Context, userID uint64, refreshHash []byte, _ time.Time, ua, ip string) {
		assert.Equal(t, userID, uint64(42))
		assert.Equal(t, len(refreshHash), 32) // sha256 == 32 bytes
		assert.Equal(t, ua, in.UserAgent)
		assert.Equal(t, ip, in.IP)
	}).Return(nil)

	info, err := s.svc.Register(ctx, in)
	assert.NilError(t, err)
	assert.Equal(t, info.UserID, uint64(42))
	assert.Assert(t, info.AccessToken != "")
	assert.Assert(t, info.RefreshToken != "")
}

func (s *RegisterSuite) TestEmailAlreadyExists() {
	t := s.T()
	ctx := t.Context()
	in := domain.RegisterInput{Email: "taken@example.com", Password: "Password123!"}

	s.authStorage.GetUserByEmailMock.Expect(ctx, in.Email).Return(&domain.User{ID: 1, Email: in.Email}, nil)

	info, err := s.svc.Register(ctx, in)
	assert.Assert(t, errors.Is(err, ErrEmailAlreadyExists))
	assert.Assert(t, info == nil)
}

func (s *RegisterSuite) TestInvalidEmailShortCircuits() {
	t := s.T()
	ctx := t.Context()
	// Invalid email must fail before any storage call — the mocks have NO
	// expectations registered; if Register calls them the test fails.
	info, err := s.svc.Register(ctx, domain.RegisterInput{Email: "not-an-email", Password: "Password123!"})
	assert.Assert(t, errors.Is(err, ErrInvalidEmail))
	assert.Assert(t, info == nil)
}

func (s *RegisterSuite) TestCreateUserStorageError() {
	t := s.T()
	ctx := t.Context()
	in := domain.RegisterInput{Email: "ok@example.com", Password: "Password123!"}

	s.authStorage.GetUserByEmailMock.Expect(ctx, in.Email).Return(nil, nil)
	storageErr := errors.New("postgres: unique violation")
	s.authStorage.CreateUserMock.Set(func(_ context.Context, _, _ string) (uint64, error) {
		return 0, storageErr
	})

	info, err := s.svc.Register(ctx, in)
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, info == nil)
}

func (s *RegisterSuite) TestCreateSessionError() {
	t := s.T()
	ctx := t.Context()
	in := domain.RegisterInput{Email: "ok@example.com", Password: "Password123!"}

	s.authStorage.GetUserByEmailMock.Expect(ctx, in.Email).Return(nil, nil)
	s.authStorage.CreateUserMock.Set(func(_ context.Context, _, _ string) (uint64, error) {
		return 7, nil
	})
	storageErr := errors.New("redis: connection refused")
	s.sessionStorage.CreateSessionMock.Set(func(_ context.Context, _ uint64, _ []byte, _ time.Time, _, _ string) error {
		return storageErr
	})

	info, err := s.svc.Register(ctx, in)
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, info == nil)
}

func TestRegisterSuite(t *testing.T) { suite.Run(t, new(RegisterSuite)) }
