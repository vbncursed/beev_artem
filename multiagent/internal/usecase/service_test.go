package usecase

import (
	"github.com/stretchr/testify/suite"

	"github.com/artem13815/hr/multiagent/internal/usecase/mocks"
)

// baseSuite gives each per-method suite a fresh MultiAgentService wired with
// a fresh minimock storage. The mock registers itself with t.Cleanup, so any
// unmet expectation fails the test automatically — no manual
// AssertExpectations.
type baseSuite struct {
	suite.Suite
	storage *mocks.DecisionStorageMock
	svc     *MultiAgentService
}

func (s *baseSuite) SetupTest() {
	t := s.T()
	s.storage = mocks.NewDecisionStorageMock(t)
	s.svc = NewMultiAgentService(s.storage)
}
