package bootstrap

import (
	"github.com/artem13815/hr/resume/internal/infrastructure/extractor"
	"github.com/artem13815/hr/resume/internal/infrastructure/persistence"
	"github.com/artem13815/hr/resume/internal/infrastructure/profile"
	"github.com/artem13815/hr/resume/internal/usecase"
)

func InitResumeService(storage *persistence.ResumeStorage) *usecase.ResumeService {
	return usecase.NewResumeService(storage, extractor.New(), profile.New())
}
