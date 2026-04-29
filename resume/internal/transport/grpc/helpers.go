package grpc

import (
	"github.com/artem13815/hr/resume/internal/domain"
	pb_models "github.com/artem13815/hr/resume/internal/pb/models"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toPBCandidate(c domain.Candidate) *pb_models.Candidate {
	return &pb_models.Candidate{
		Id:        c.ID,
		VacancyId: c.VacancyID,
		FullName:  c.FullName,
		Email:     c.Email,
		Phone:     c.Phone,
		Source:    c.Source,
		Comment:   c.Comment,
		CreatedAt: timestamppb.New(c.CreatedAt),
	}
}

func toPBResume(r domain.Resume) *pb_models.Resume {
	return &pb_models.Resume{
		Id:            r.ID,
		CandidateId:   r.CandidateID,
		FileName:      r.FileName,
		FileType:      r.FileType,
		FileSizeBytes: r.FileSizeBytes,
		StoragePath:   r.StoragePath,
		ExtractedText: r.ExtractedText,
		CreatedAt:     timestamppb.New(r.CreatedAt),
	}
}
