package auth_service

import (
	"time"

	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"

	"github.com/artem13815/hr/auth/internal/services/auth_service/mocks"
)

// 32 bytes — same minimum the production validate() enforces.
const testJWTSecret = "test-jwt-secret-must-be-32-bytes!"

const (
	testAccessTTL  = time.Minute
	testRefreshTTL = time.Hour

	// MinCost keeps bcrypt fast in tests; production uses DefaultCost (10) or
	// whatever the operator sets in AuthConfig.BcryptCost.
	testBcryptCost = bcrypt.MinCost
)

// baseSuite gives each per-method suite a fresh AuthService wired with fresh
// minimock mocks. Mocks register themselves with t.Cleanup, so any unmet
// expectation fails the test automatically — no manual AssertExpectations.
type baseSuite struct {
	suite.Suite
	authStorage    *mocks.AuthStorageMock
	sessionStorage *mocks.SessionStorageMock
	svc            *AuthService
}

func (s *baseSuite) SetupTest() {
	t := s.T()
	s.authStorage = mocks.NewAuthStorageMock(t)
	s.sessionStorage = mocks.NewSessionStorageMock(t)
	s.svc = NewAuthService(
		s.authStorage,
		s.sessionStorage,
		testJWTSecret,
		testAccessTTL,
		testRefreshTTL,
		testBcryptCost,
	)
}
