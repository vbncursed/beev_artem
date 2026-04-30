package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/analysis/internal/domain"
)

type GetAnalysisSuite struct{ baseSuite }

func (s *GetAnalysisSuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	want := &domain.Analysis{ID: "a-1", VacancyID: "v-1", CandidateID: "c-1"}

	s.storage.GetAnalysisMock.Expect(ctx, "a-1", uint64(7), false).Return(want, nil)

	got, err := s.svc.GetAnalysis(ctx, domain.GetAnalysisInput{
		RequestUserID: 7,
		AnalysisID:    "a-1",
	})
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

func (s *GetAnalysisSuite) TestSuccessAdmin() {
	t := s.T()
	ctx := t.Context()
	want := &domain.Analysis{ID: "a-1"}

	s.storage.GetAnalysisMock.Expect(ctx, "a-1", uint64(7), true).Return(want, nil)

	got, err := s.svc.GetAnalysis(ctx, domain.GetAnalysisInput{
		RequestUserID: 7,
		IsAdmin:       true,
		AnalysisID:    "a-1",
	})
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

func (s *GetAnalysisSuite) TestInvalidArgumentZeroUser() {
	t := s.T()
	got, err := s.svc.GetAnalysis(t.Context(), domain.GetAnalysisInput{AnalysisID: "a-1"})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

func (s *GetAnalysisSuite) TestInvalidArgumentEmptyID() {
	t := s.T()
	got, err := s.svc.GetAnalysis(t.Context(), domain.GetAnalysisInput{
		RequestUserID: 1,
		AnalysisID:    "   ",
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, got == nil)
}

// TestNotFoundOnNilStorageResult: storage returns (nil, nil) when ownership
// filter excludes the row — service must translate that to ErrNotFound
// rather than leak nil.
func (s *GetAnalysisSuite) TestNotFoundOnNilStorageResult() {
	t := s.T()
	ctx := t.Context()
	s.storage.GetAnalysisMock.Expect(ctx, "missing", uint64(1), false).Return(nil, nil)

	got, err := s.svc.GetAnalysis(ctx, domain.GetAnalysisInput{
		RequestUserID: 1,
		AnalysisID:    "missing",
	})
	assert.ErrorIs(t, err, ErrNotFound)
	assert.Assert(t, got == nil)
}

func (s *GetAnalysisSuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	storageErr := errors.New("pgx: connection refused")

	s.storage.GetAnalysisMock.Expect(ctx, "a-1", uint64(1), false).Return(nil, storageErr)

	got, err := s.svc.GetAnalysis(ctx, domain.GetAnalysisInput{
		RequestUserID: 1,
		AnalysisID:    "a-1",
	})
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, got == nil)
}

func TestGetAnalysisSuite(t *testing.T) { suite.Run(t, new(GetAnalysisSuite)) }
