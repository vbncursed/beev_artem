package bootstrap

import (
	"github.com/artem13815/hr/resume/internal/services/resume_service"
	"github.com/artem13815/hr/resume/internal/storage/resume_storage"
)

func InitResumeService(storage *resume_storage.ResumeStorage) *resume_service.ResumeService {
	return resume_service.NewResumeService(storage)
}
