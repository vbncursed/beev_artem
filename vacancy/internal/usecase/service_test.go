package usecase

import (
	"github.com/stretchr/testify/suite"

	"github.com/artem13815/hr/vacancy/internal/usecase/mocks"
)

// baseSuite gives each per-method suite a fresh VacancyService wired with a
// fresh minimock storage. The mock registers itself with t.Cleanup, so any
// unmet expectation fails the test automatically — no manual
// AssertExpectations.
type baseSuite struct {
	suite.Suite
	storage *mocks.VacancyStorageMock
	svc     *VacancyService
}

func (s *baseSuite) SetupTest() {
	t := s.T()
	s.storage = mocks.NewVacancyStorageMock(t)
	s.svc = NewVacancyService(s.storage)
}
