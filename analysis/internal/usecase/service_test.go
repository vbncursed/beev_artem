package usecase

import (
	"context"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"

	pb_multiagent "github.com/artem13815/hr/analysis/internal/pb/multiagent_api"
	"github.com/artem13815/hr/analysis/internal/usecase/mocks"
)

// baseSuite gives each per-method suite a fresh AnalysisService wired with
// fresh minimock storage + scorer. The mocks register themselves with
// t.Cleanup, so any unmet expectation fails the test automatically.
//
// multiAgentClient defaults to nil — most tests don't go through the LLM
// fan-out branch (`s.multiAgentClient != nil` short-circuits). Tests that
// DO need to exercise the fan-out construct their own AnalysisService with
// a stub client, see start_test.go.
type baseSuite struct {
	suite.Suite
	storage *mocks.AnalysisStorageMock
	scorer  *mocks.ScorerMock
	svc     *AnalysisService
}

func (s *baseSuite) SetupTest() {
	t := s.T()
	s.storage = mocks.NewAnalysisStorageMock(t)
	s.scorer = mocks.NewScorerMock(t)
	s.svc = NewAnalysisService(s.storage, s.scorer, nil)
}

// multiagentClientStub is a hand-rolled minimal stub for the pb-generated
// MultiAgentServiceClient. We implement only the one method usecase calls
// (GenerateDecision) and capture the call so tests can assert what was sent
// and inject a canned response or error.
//
// Going through pb.MultiAgentServiceClient (as opposed to a narrower local
// interface) keeps the test honest: the production code dials this exact
// interface, so signature drift surfaces here first.
type multiagentClientStub struct {
	resp        *pb_multiagent.GenerateDecisionResponse
	err         error
	called      bool
	capturedReq *pb_multiagent.GenerateDecisionRequest
}

func (m *multiagentClientStub) GenerateDecision(_ context.Context, in *pb_multiagent.GenerateDecisionRequest, _ ...grpc.CallOption) (*pb_multiagent.GenerateDecisionResponse, error) {
	m.called = true
	m.capturedReq = in
	return m.resp, m.err
}

// ClassifyRole is required by pb.MultiAgentServiceClient but never invoked
// by analysis usecase (only vacancy calls it). The stub keeps the interface
// satisfied; if analysis ever starts using it, tests will need to assert on
// the captured request like GenerateDecision does.
func (m *multiagentClientStub) ClassifyRole(_ context.Context, _ *pb_multiagent.ClassifyRoleRequest, _ ...grpc.CallOption) (*pb_multiagent.ClassifyRoleResponse, error) {
	return nil, nil
}
