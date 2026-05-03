package usecase

import (
	"github.com/stretchr/testify/suite"

	"github.com/artem13815/hr/admin/internal/usecase/mocks"
)

// baseSuite gives each per-method suite a fresh AdminService wired with fresh
// minimock collaborators. The mocks register themselves with t.Cleanup so any
// unmet expectation fails the test automatically — no manual
// AssertExpectations.
type baseSuite struct {
	suite.Suite
	storage    *mocks.AdminStorageMock
	authClient *mocks.AuthClientMock
	svc        *AdminService
}

func (s *baseSuite) SetupTest() {
	t := s.T()
	s.storage = mocks.NewAdminStorageMock(t)
	s.authClient = mocks.NewAuthClientMock(t)
	s.svc = NewAdminService(s.storage, s.authClient)
}
