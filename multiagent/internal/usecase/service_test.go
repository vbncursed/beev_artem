package usecase

import (
	"github.com/stretchr/testify/suite"

	"github.com/artem13815/hr/multiagent/internal/usecase/mocks"
)

// baseSuite gives each per-method suite a fresh MultiAgentService wired
// with fresh minimock storage / llm / prompt-store. Mocks register
// themselves with t.Cleanup so any unmet expectation fails the test
// automatically — no manual AssertExpectations.
type baseSuite struct {
	suite.Suite
	storage *mocks.DecisionStorageMock
	llm     *mocks.LLMMock
	prompts *mocks.PromptStoreMock
	svc     *MultiAgentService
}

func (s *baseSuite) SetupTest() {
	t := s.T()
	s.storage = mocks.NewDecisionStorageMock(t)
	s.llm = mocks.NewLLMMock(t)
	s.prompts = mocks.NewPromptStoreMock(t)
	s.svc = NewMultiAgentService(s.storage, s.llm, s.prompts)
}
