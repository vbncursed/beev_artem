package resume_service

import (
	"context"

	"github.com/artem13815/hr/resume/internal/domain"
)

const maxBatchFiles = 50

func (s *ResumeService) IngestResumeBatch(ctx context.Context, in domain.BatchIngestResumeInput) (*domain.BatchIngestResumeResult, error) {
	if in.RequestUserID == 0 || len(in.Files) == 0 || len(in.Files) > maxBatchFiles {
		return nil, ErrInvalidArgument
	}

	results := make([]domain.BatchIngestResumeItemResult, 0, len(in.Files))
	for i, file := range in.Files {
		externalID := file.ExternalID
		if externalID == "" {
			externalID = "item-" + itoa(i+1)
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

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for v > 0 {
		i--
		buf[i] = byte('0' + (v % 10))
		v /= 10
	}
	return string(buf[i:])
}
