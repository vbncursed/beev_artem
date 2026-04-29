package usecase

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"
	"gotest.tools/v3/assert"

	"github.com/artem13815/hr/resume/internal/domain"
)

type IngestResumeBatchSuite struct{ baseSuite }

func (s *IngestResumeBatchSuite) TestSuccessAllFilesProcessed() {
	t := s.T()
	ctx := t.Context()

	// Each successful file results in one storage call returning a candidate.
	calls := 0
	s.storage.CreateCandidateWithResumeMock.Set(func(_ context.Context, _ domain.CreateCandidateInput, _ domain.NewResumeData) (*domain.Candidate, *domain.Resume, error) {
		calls++
		return &domain.Candidate{ID: "c"}, &domain.Resume{ID: "r"}, nil
	})

	in := domain.BatchIngestResumeInput{
		RequestUserID: 1,
		Files: []domain.ResumeIntakeFile{
			{ExternalID: "ext-a", FileData: resumeText},
			{ExternalID: "ext-b", FileData: resumeText},
		},
	}

	out, err := s.svc.IngestResumeBatch(ctx, in)
	assert.NilError(t, err)
	assert.Equal(t, len(out.Results), 2)
	assert.Equal(t, calls, 2)
	assert.Equal(t, out.Results[0].ExternalID, "ext-a")
	assert.Equal(t, out.Results[1].ExternalID, "ext-b")
	assert.Equal(t, out.Results[0].Error, "")
	assert.Equal(t, out.Results[1].Error, "")
}

func (s *IngestResumeBatchSuite) TestPartialFailureRecorded() {
	t := s.T()
	ctx := t.Context()

	in := domain.BatchIngestResumeInput{
		RequestUserID: 1,
		Files: []domain.ResumeIntakeFile{
			// First file is unparseable garbage — extractor rejects it before
			// storage is touched.
			{ExternalID: "bad", FileData: []byte{0x00, 0x01}},
			// Second file is valid TXT.
			{ExternalID: "good", FileData: resumeText},
		},
	}

	s.storage.CreateCandidateWithResumeMock.Set(func(_ context.Context, _ domain.CreateCandidateInput, _ domain.NewResumeData) (*domain.Candidate, *domain.Resume, error) {
		return &domain.Candidate{ID: "c"}, &domain.Resume{ID: "r"}, nil
	})

	out, err := s.svc.IngestResumeBatch(ctx, in)
	assert.NilError(t, err)
	assert.Equal(t, len(out.Results), 2)

	assert.Equal(t, out.Results[0].ExternalID, "bad")
	assert.Assert(t, out.Results[0].Error != "")
	assert.Assert(t, out.Results[0].Candidate == nil)

	assert.Equal(t, out.Results[1].ExternalID, "good")
	assert.Equal(t, out.Results[1].Error, "")
	assert.Assert(t, out.Results[1].Candidate != nil)
}

func (s *IngestResumeBatchSuite) TestExternalIDFallback() {
	t := s.T()
	ctx := t.Context()

	s.storage.CreateCandidateWithResumeMock.Set(func(_ context.Context, _ domain.CreateCandidateInput, _ domain.NewResumeData) (*domain.Candidate, *domain.Resume, error) {
		return &domain.Candidate{ID: "c"}, &domain.Resume{ID: "r"}, nil
	})

	in := domain.BatchIngestResumeInput{
		RequestUserID: 1,
		Files: []domain.ResumeIntakeFile{
			{FileData: resumeText}, // empty ExternalID → service synthesizes "item-1"
			{FileData: resumeText},
		},
	}

	out, err := s.svc.IngestResumeBatch(ctx, in)
	assert.NilError(t, err)
	assert.Equal(t, out.Results[0].ExternalID, "item-1")
	assert.Equal(t, out.Results[1].ExternalID, "item-2")
}

func (s *IngestResumeBatchSuite) TestInvalidArgumentZeroUser() {
	t := s.T()
	out, err := s.svc.IngestResumeBatch(t.Context(), domain.BatchIngestResumeInput{
		Files: []domain.ResumeIntakeFile{{FileData: resumeText}},
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, out == nil)
}

func (s *IngestResumeBatchSuite) TestInvalidArgumentEmptyBatch() {
	t := s.T()
	out, err := s.svc.IngestResumeBatch(t.Context(), domain.BatchIngestResumeInput{
		RequestUserID: 1,
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, out == nil)
}

func (s *IngestResumeBatchSuite) TestInvalidArgumentBatchTooLarge() {
	t := s.T()
	files := make([]domain.ResumeIntakeFile, 51)
	out, err := s.svc.IngestResumeBatch(t.Context(), domain.BatchIngestResumeInput{
		RequestUserID: 1,
		Files:         files,
	})
	assert.ErrorIs(t, err, ErrInvalidArgument)
	assert.Assert(t, out == nil)
}

func TestIngestResumeBatchSuite(t *testing.T) {
	suite.Run(t, new(IngestResumeBatchSuite))
}
