package usecase

import (
	"context"
	"strconv"

	"github.com/artem13815/hr/resume/internal/domain"
)

// MaxBatchFiles caps a single IngestResumeBatch call. Combined with
// MaxResumeSizeBytes (10 MB per file) and gRPC MaxRecvMsgSize (~11 MB), the
// realistic batch size is bounded by the transport — this constant is the
// per-RPC business cap so a caller cannot ask the service to chew through
// 50 large files in one shot even when files are tiny. Exported so the API
// layer can fail-fast before the service is called.
const MaxBatchFiles = 50

func (s *ResumeService) IngestResumeBatch(ctx context.Context, in domain.BatchIngestResumeInput) (*domain.BatchIngestResumeResult, error) {
	if in.RequestUserID == 0 || len(in.Files) == 0 || len(in.Files) > MaxBatchFiles {
		return nil, ErrInvalidArgument
	}

	results := make([]domain.BatchIngestResumeItemResult, 0, len(in.Files))
	for i, file := range in.Files {
		externalID := file.ExternalID
		if externalID == "" {
			externalID = "item-" + strconv.Itoa(i+1)
		}

		one, err := s.CreateCandidateFromResume(ctx, domain.CreateCandidateFromResumeInput{
			RequestUserID: in.RequestUserID,
			VacancyID:     "",
			FileData:      file.FileData,
		})
		if err != nil {
			results = append(results, domain.BatchIngestResumeItemResult{ExternalID: externalID, Error: err.Error()})
			continue
		}

		results = append(results, domain.BatchIngestResumeItemResult{
			ExternalID: externalID,
			Candidate:  one.Candidate,
			Resume:     one.Resume,
		})
	}

	return &domain.BatchIngestResumeResult{Results: results}, nil
}
