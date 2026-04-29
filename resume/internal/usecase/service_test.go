package usecase

import (
	"github.com/stretchr/testify/suite"

	"github.com/artem13815/hr/resume/internal/infrastructure/extractor"
	"github.com/artem13815/hr/resume/internal/infrastructure/profile"
	"github.com/artem13815/hr/resume/internal/usecase/mocks"
)

// baseSuite gives each per-method suite a fresh ResumeService wired with a
// fresh minimock storage. The mock registers itself with t.Cleanup, so any
// unmet expectation fails the test automatically — no manual
// AssertExpectations.
//
// Extractor / profile ports are wired with the *real* implementations. They
// are pure (no I/O) so tests stay fast and deterministic; mocking them would
// just duplicate the production code in test fixtures. Tests that need a
// failure path use minimock-overridable variants where it matters (none so
// far).
type baseSuite struct {
	suite.Suite
	storage *mocks.ResumeStorageMock
	svc     *ResumeService
}

func (s *baseSuite) SetupTest() {
	t := s.T()
	s.storage = mocks.NewResumeStorageMock(t)
	s.svc = NewResumeService(s.storage, extractor.New(), profile.New())
}
