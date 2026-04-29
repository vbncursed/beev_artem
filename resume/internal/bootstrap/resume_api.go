package bootstrap

import (
	"github.com/artem13815/hr/resume/internal/api/resume_service_api"
	"github.com/artem13815/hr/resume/internal/services/resume_service"
)

func InitResumeServiceAPI(service *resume_service.ResumeService) *resume_service_api.ResumeServiceAPI {
	return resume_service_api.NewResumeServiceAPI(service)
}
