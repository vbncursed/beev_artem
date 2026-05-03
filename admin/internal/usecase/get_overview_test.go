package usecase

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/admin/internal/domain"
)

type GetOverviewSuite struct{ baseSuite }

func (s *GetOverviewSuite) TestSuccess() {
	t := s.T()
	ctx := t.Context()
	want := &domain.SystemStats{
		UsersTotal:      42,
		AdminsTotal:     3,
		VacanciesTotal:  100,
		CandidatesTotal: 250,
		AnalysesTotal:   500,
		AnalysesDone:    480,
		AnalysesFailed:  5,
	}

	s.storage.GetSystemStatsMock.Expect(ctx).Return(want, nil)

	got, err := s.svc.GetOverview(ctx)
	assert.NilError(t, err)
	assert.DeepEqual(t, got, want)
}

func (s *GetOverviewSuite) TestStorageError() {
	t := s.T()
	ctx := t.Context()
	storageErr := errors.New("pgx: connection refused")

	s.storage.GetSystemStatsMock.Expect(ctx).Return(nil, storageErr)

	got, err := s.svc.GetOverview(ctx)
	assert.ErrorIs(t, err, storageErr)
	assert.Assert(t, got == nil)
}

func TestGetOverviewSuite(t *testing.T) { suite.Run(t, new(GetOverviewSuite)) }
