package resume_service_api

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/artem13815/hr/resume/internal/domain"
	pb_models "github.com/artem13815/hr/resume/internal/pb/models"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type userContext struct {
	UserID  uint64
	Role    string
	IsAdmin bool
}

func getUserContext(ctx context.Context) (*userContext, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("metadata not found")
	}

	role := "user"
	if values := md.Get("x-user-role"); len(values) > 0 && strings.TrimSpace(values[0]) != "" {
		role = strings.ToLower(strings.TrimSpace(values[0]))
	}

	keys := []string{"x-user-id", "user-id"}
	for _, key := range keys {
		if values := md.Get(key); len(values) > 0 {
			userID, err := strconv.ParseUint(values[0], 10, 64)
			if err != nil || userID == 0 {
				return nil, fmt.Errorf("invalid user id")
			}
			return &userContext{UserID: userID, Role: role, IsAdmin: role == "admin"}, nil
		}
	}

	return nil, fmt.Errorf("user id not found")
}

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
