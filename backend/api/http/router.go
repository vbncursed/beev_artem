package http

import (
	"github.com/gofiber/fiber/v2"

	"github.com/artem13815/hr/api/http/handlers"
)

// Register wires all HTTP routes onto given Fiber app.
func Register(app *fiber.App, auth *handlers.AuthHandler, health *handlers.HealthHandler, resume *handlers.ResumeHandler, vacancy *handlers.VacancyHandler, authRequired fiber.Handler, resumes *handlers.ResumesHandler, analyses *handlers.AnalysisHandler) {
	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Health and readiness endpoints for probes/monitoring
	v1.Get("/health", health.Health)
	v1.Get("/ready", health.Ready)

	a := v1.Group("/auth")
	a.Post("/register", auth.Register)
	a.Post("/login", auth.Login)

	// Resume analysis
	rg := v1.Group("/resume", authRequired)
	rg.Post("/analyze", resume.Analyze)

	// Resumes storage
	rs := v1.Group("/resumes", authRequired)
	rs.Post("/", resumes.Upload)
	rs.Get("/", resumes.List)
	rs.Get("/:id", resumes.Get)
	rs.Get("/:id/profile", resumes.Profile)
	rs.Post("/:id/profile/rebuild", resumes.RebuildProfile)
	rs.Get("/:id/file", resumes.Download)
	rs.Delete("/:id", resumes.Delete)

	// Analyses
	ag := v1.Group("/analyses", authRequired)
	ag.Post("/", analyses.Create)
	ag.Get("/", analyses.List)
	ag.Get("/:id", analyses.Get)
	ag.Delete("/:id", analyses.Delete)

	// Vacancies
	vg := v1.Group("/vacancies", authRequired)
	vg.Post("/", vacancy.Create)
	vg.Get("/", vacancy.List)
	vg.Get("/:id", vacancy.GetByID)
	vg.Put("/:id/skills", vacancy.UpdateSkills)
	vg.Delete("/:id", vacancy.Delete)
	vg.Get("/:id/analyses", analyses.ListByVacancy)
	vg.Get("/:id/candidates", analyses.CandidatesByVacancy)
}
