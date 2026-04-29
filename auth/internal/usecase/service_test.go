package usecase

import (
	"time"

	"github.com/stretchr/testify/suite"
	"golang.org/x/crypto/bcrypt"

	"github.com/artem13815/hr/auth/internal/infrastructure/jwt"
	"github.com/artem13815/hr/auth/internal/usecase/mocks"
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
// minimock mocks for the storage ports and the *real* jwt.Issuer for the
// TokenIssuer port. The issuer is pure (no I/O), so using the real
// implementation keeps tests honest about the production wire format
// without slowing them down.
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
		jwt.NewIssuer(testJWTSecret, testAccessTTL),
		testRefreshTTL,
		testBcryptCost,
	)
}
